package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"app/internal/config"
	"app/internal/http/handlers"
	"app/internal/http/middleware"
	"app/internal/store/postgres"
)

func main() {
	cfg := config.Load()

	store, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	authHandler := handlers.NewAuthHandler(store, cfg.JWTSecret)
	settingsHandler := handlers.NewSettingsHandler(store)
	balancesHandler := handlers.NewBalancesHandler()

	// Public routes
	router.GET("/api/health", handlers.HealthzHandler)
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/logout", authHandler.Logout)

	// Protected routes
	auth := router.Group("/api")
	auth.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		auth.GET("/auth/me", authHandler.Me)
		auth.GET("/balances", balancesHandler.GetBalances)
		auth.PATCH("/me/settings", settingsHandler.UpdateSettings)
	}

	addr := ":" + cfg.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
