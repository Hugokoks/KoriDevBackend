package ratelimiter

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/ratelimit"
)

const SessionKey = "rate_limit_id"
const CleanupInterval = 30 * time.Minute // Úklid spustit každých 30 minut
const MaxInactivity = 1 * time.Hour

// LimiterEntry uchovává limiter a čas jeho posledního použití pro účely úklidu.
type LimiterEntry struct {
	Limiter  ratelimit.Limiter
	LastSeen time.Time // Čas poslední aktivity
}

// limiterStore uchovává LimiterEntry pro každé Session ID.
var limiterStore sync.Map // Klíč: string (Session ID), Hodnota: LimiterEntry

// StartCleanupRoutine spouští gorutinu pro mazání starých, nepoužívaných limiterů.
func StartCleanupRoutine() {
	go func() {
		for {
			time.Sleep(CleanupInterval)

			// Procházení celé sync.Map
			limiterStore.Range(func(key, value interface{}) bool {
				entry, ok := value.(LimiterEntry)
				if !ok {
					return true
				}

				// Smaž, pokud je neaktivní déle než MaxInactivity
				if time.Since(entry.LastSeen) > MaxInactivity {
					limiterStore.Delete(key)
				}
				return true
			})
		}
	}()
}

// AnonymousRateLimiter je middleware, které zajišťuje Session ID a aplikuje Rate Limiting.
func AnonymousRateLimiter() gin.HandlerFunc {

	// Přednastavený limit: 1 požadavek každé 2 sekundy (efektivně 5 za 10s)
	limitDuration := time.Second * 2

	return func(c *gin.Context) {
		session := sessions.Default(c)

		// 1. Zajištění unikátního Session ID
		sessionID := session.Get(SessionKey)

		if sessionID == nil {
			newID := uuid.New().String()
			session.Set(SessionKey, newID)

			session.Options(sessions.Options{
				MaxAge:   30 * 24 * 3600, // Session cookie platí 30 dní
				Path:     "/",
				HttpOnly: true,
			})
			session.Save()

			sessionID = newID
		}

		limiterKey := sessionID.(string)

		// 2. Získání nebo vytvoření LimiterEntry
		limiterValue, loaded := limiterStore.Load(limiterKey)

		var entry LimiterEntry

		if !loaded {
			// Vytvoříme nový entry
			entry = LimiterEntry{
				Limiter:  ratelimit.New(1, ratelimit.Per(limitDuration)),
				LastSeen: time.Now(),
			}
			limiterStore.Store(limiterKey, entry)
		} else {
			entry = limiterValue.(LimiterEntry)
			// Aktualizujeme čas LastSeen pro úklidovou rutinu
			entry.LastSeen = time.Now()
			// Pozn.: Zde ukládáme entry zpět, aby se aktualizoval čas LastSeen PŘED kontrolou Take().
			limiterStore.Store(limiterKey, entry)
		}

		rl := entry.Limiter // Získáme samotný limiter

		// 3. Aplikace Rate Limitingu
		now := time.Now()
		next := rl.Take() // Spotřebuje token a vrátí čas, kdy bude k dispozici další

		// Kontrola, zda nás fronta neposlala dále, než je základní interval (limitDuration)
		if next.After(now.Add(limitDuration)) {

			timeToWait := next.Sub(now) // Doba, kterou musí uživatel čekat

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":          "Rate limit překročen",
				"retry_after_ms": timeToWait.Milliseconds(),
				"message":        "Zkuste to prosím znovu za " + timeToWait.Round(time.Millisecond).String() + ".",
			})

			// Uložíme modifikovaný stav limiteru zpět do mapy
			limiterStore.Store(limiterKey, entry)
			return
		}

		// Uložíme modifikovaný stav limiteru zpět do mapy po úspěšném průchodu
		limiterStore.Store(limiterKey, entry)

		// 4. Limit NEBYL překročen: Pokračovat
		c.Next()
	}
}
