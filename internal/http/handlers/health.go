package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type BalancesHandler struct{}

func NewBalancesHandler() *BalancesHandler {
	return &BalancesHandler{}
}

func (h *BalancesHandler) GetBalances(c *gin.Context) {
	// TODO: implement balances calculation
	c.JSON(http.StatusOK, gin.H{
		"balances": gin.H{
			"total_minor": 0,
		},
	})
}
