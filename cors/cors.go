package cors

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupCors konfiguruje a vrací CORS middleware.
// Povolená doména se načítá z proměnné prostředí CLIENT_URL.
func SetupCors() gin.HandlerFunc {
	// 1. Načtení povolené URL frontendu z proměnné prostředí
	clientURL := os.Getenv("CLIENT_URL")
	if clientURL == "" {
		// Nastavení defaultu pro lokální testování, pokud ENV proměnná chybí
		clientURL = "http://localhost:5173"
		println("Upozornění CORS: Proměnná CLIENT_URL nebyla nastavena. Používám default: " + clientURL)
	}

	// 2. Konfigurace CORS
	config := cors.DefaultConfig()

	// Povolí pouze jednu konkrétní doménu, kterou jsme načetli z ENV
	config.AllowOrigins = []string{clientURL}

	// Povolí metody
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	// Povolí hlavičky (nutné pro Authorization, Content-Type apod.)
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}

	// Povolí odesílání cookies (pokud je to potřeba pro sezení/autentizaci)
	// config.AllowCredentials = true

	// Nastaví maximální dobu cachování preflight požadavků (12 hodin)
	config.MaxAge = 12 * time.Hour

	// 3. Vrací nakonfigurovaný CORS middleware
	return cors.New(config)
}
