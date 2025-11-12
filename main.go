package main

import (
	"net/http"

	"koridev/cors"
	"koridev/ratelimiter"
	"koridev/router"
	"koridev/sessions"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	/////ratelimiter map cleanup
	ratelimiter.StartCleanupRoutine()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// ----- CORS -----
	r.Use(cors.SetupCors())
	r.SetTrustedProxies(nil)

	// ---- Security hlavičky ----

	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		// HSTS zapni jen pokud jsi za HTTPS reverse proxy:
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Next()
	})
	// ---- Limit velikosti request body ----

	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*1024*1024) // 2 MB
		c.Next()
	})

	r.Use(sessions.SetupSessions())           // musí nastavit Secure/HttpOnly atd.
	r.Use(ratelimiter.AnonymousRateLimiter()) // limiter používající sessionID

	api := r.Group("/api")

	api.OPTIONS("/*path", func(c *gin.Context) { // ← hook pro preflight
		c.Status(http.StatusNoContent)
	})

	/////ROUTES
	
	// ===== HEALTHCHECK ENDPOINT =====
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Backend is healthy 🚀",
		})
	})

	router.Message(api)

	r.Run()
}
