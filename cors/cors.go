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
	clientURL := os.Getenv("CLIENT_URL")
	if clientURL == "" {
		clientURL = "http://localhost:5173"
		println("⚠️  CLIENT_URL nebyla nastavena, používám default:", clientURL)
	}

	config := cors.Config{
		AllowOrigins:     []string{clientURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // 🔥 DŮLEŽITÉ – bez toho browser nepošle skutečný POST
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}