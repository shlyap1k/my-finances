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
	fmt.Println(cfg.DatabaseURL)
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
	incomeRuleHandler := handlers.NewIncomeRuleHandler(store)
	expenseRuleHandler := handlers.NewExpenseRuleHandler(store)
	savingsAccountHandler := handlers.NewSavingsAccountHandler(store)
	savingsRuleHandler := handlers.NewSavingsRuleHandler(store)

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
		
		// Income rules routes
		auth.GET("/income-rules", incomeRuleHandler.ListIncomeRules)
		auth.POST("/income-rules", incomeRuleHandler.CreateIncomeRule)
		auth.GET("/income-rules/:id", incomeRuleHandler.GetIncomeRule)
		auth.PUT("/income-rules/:id", incomeRuleHandler.UpdateIncomeRule)
		auth.DELETE("/income-rules/:id", incomeRuleHandler.DeleteIncomeRule)

		// Expense rules routes
		auth.GET("/expense-rules", expenseRuleHandler.ListExpenseRules)
		auth.POST("/expense-rules", expenseRuleHandler.CreateExpenseRule)
		auth.GET("/expense-rules/:id", expenseRuleHandler.GetExpenseRule)
		auth.PUT("/expense-rules/:id", expenseRuleHandler.UpdateExpenseRule)
		auth.DELETE("/expense-rules/:id", expenseRuleHandler.DeleteExpenseRule)

		// Savings accounts routes
		auth.GET("/savings-accounts", savingsAccountHandler.ListSavingsAccounts)
		auth.POST("/savings-accounts", savingsAccountHandler.CreateSavingsAccount)
		auth.GET("/savings-accounts/:id", savingsAccountHandler.GetSavingsAccount)
		auth.PUT("/savings-accounts/:id", savingsAccountHandler.UpdateSavingsAccount)
		auth.DELETE("/savings-accounts/:id", savingsAccountHandler.DeleteSavingsAccount)

		// Savings rules routes
		auth.GET("/savings-rules", savingsRuleHandler.ListSavingsRules)
		auth.POST("/savings-rules", savingsRuleHandler.CreateSavingsRule)
		auth.GET("/savings-rules/:id", savingsRuleHandler.GetSavingsRule)
		auth.PUT("/savings-rules/:id", savingsRuleHandler.UpdateSavingsRule)
		auth.DELETE("/savings-rules/:id", savingsRuleHandler.DeleteSavingsRule)
	}

	addr := ":" + cfg.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
