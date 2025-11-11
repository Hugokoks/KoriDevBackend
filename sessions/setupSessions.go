package sessions

import (
	"log" // Pro logování chyby
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// SetupSession inicializuje cookie-based sessions middleware, tajný klíč tahá z ENV
func SetupSessions() gin.HandlerFunc {
	// 1. Bezpečné načtení tajného klíče
	secretKey := os.Getenv("SESSION_SECRET")

	// Důležité: Kontrola, zda byl klíč nastaven
	if secretKey == "" {
		log.Fatal("Chyba: Proměnná prostředí 'SESSION_SECRET' musí být nastavena!")
		// V produkci byste měli aplikaci ukončit, pokud klíč chybí.
	}

	// 2. Vytvoření Session Store (Cookie Store)
	// Klíč musí být konvertován na []byte
	store := cookie.NewStore([]byte(secretKey))

	// 3. Aplikace sessions middleware na router
	return sessions.Sessions("anon_session", store)
}
