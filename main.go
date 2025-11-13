package main

import (
	"net/http"
	"os"

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
	/////debuging cors
	r.Use(func(c *gin.Context) {
    origin := c.Request.Header.Get("Origin")
    method := c.Request.Method
    path := c.Request.URL.Path

    if origin != "" {
        println("🌍 Incoming request:",
            "Origin =", origin,
            "| Method =", method,
            "| Path =", path,
        )
    }

    c.Next()
	})
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

	port := os.Getenv("PORT")
	if port == "" {
	    port = "8080"
	}
	println("✅ Starting server on port:", port)
	if err := r.Run(":" + port); err != nil {
	    panic(err)
	}
}
