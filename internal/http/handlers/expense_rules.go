package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"app/internal/http/dto"
	"app/internal/store/postgres"
)

type ExpenseRuleHandler struct {
	store *postgres.Store
}

func NewExpenseRuleHandler(store *postgres.Store) *ExpenseRuleHandler {
	return &ExpenseRuleHandler{
		store: store,
	}
}

type CreateExpenseRuleRequest struct {
	Name        string  `json:"name" binding:"required"`
	AmountMinor int64   `json:"amount_minor" binding:"required,min=1"`
	MonthlyDay  int     `json:"monthly_day" binding:"required,min=1,max=31"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     *string `json:"end_date,omitempty"`
	Active      bool    `json:"active"`
	Note        *string `json:"note,omitempty"`
}

type UpdateExpenseRuleRequest struct {
	Name        string  `json:"name" binding:"required"`
	AmountMinor int64   `json:"amount_minor" binding:"required,min=1"`
	MonthlyDay  int     `json:"monthly_day" binding:"required,min=1,max=31"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     *string `json:"end_date,omitempty"`
	Active      bool    `json:"active"`
	Note        *string `json:"note,omitempty"`
}

func (h *ExpenseRuleHandler) ListExpenseRules(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "not authenticated",
			},
		})
		return
	}

	var activeFilter *bool
	activeParam := c.Query("active")
	if activeParam != "" {
		if activeParam == "true" {
			val := true
			activeFilter = &val
		} else if activeParam == "false" {
			val := false
			activeFilter = &val
		}
	}

	rules, err := h.store.GetExpenseRules(c.Request.Context(), userID.(int64), activeFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get expense rules",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": rules,
	})
}

func (h *ExpenseRuleHandler) GetExpenseRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "not authenticated",
			},
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "validation",
				Message: "invalid id",
				Fields:  map[string]string{"id": "id must be a number"},
			},
		})
		return
	}

	rule, err := h.store.GetExpenseRuleByID(c.Request.Context(), userID.(int64), id)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "expense rule not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get expense rule",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expense_rule": rule,
	})
}

func (h *ExpenseRuleHandler) CreateExpenseRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "not authenticated",
			},
		})
		return
	}

	var req CreateExpenseRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "validation",
				Message: "invalid request",
				Fields:  map[string]string{"error": err.Error()},
			},
		})
		return
	}

	rule, err := h.store.CreateExpenseRule(
		c.Request.Context(),
		userID.(int64),
		req.Name,
		req.AmountMinor,
		req.MonthlyDay,
		req.StartDate,
		req.EndDate,
		req.Active,
		req.Note,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to create expense rule",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"expense_rule": rule,
	})
}

func (h *ExpenseRuleHandler) UpdateExpenseRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "not authenticated",
			},
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "validation",
				Message: "invalid id",
				Fields:  map[string]string{"id": "id must be a number"},
			},
		})
		return
	}

	var req UpdateExpenseRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "validation",
				Message: "invalid request",
				Fields:  map[string]string{"error": err.Error()},
			},
		})
		return
	}

	rule, err := h.store.UpdateExpenseRule(
		c.Request.Context(),
		userID.(int64),
		id,
		req.Name,
		req.AmountMinor,
		req.MonthlyDay,
		req.StartDate,
		req.EndDate,
		req.Active,
		req.Note,
	)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "expense rule not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to update expense rule",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expense_rule": rule,
	})
}

func (h *ExpenseRuleHandler) DeleteExpenseRule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "not authenticated",
			},
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "validation",
				Message: "invalid id",
				Fields:  map[string]string{"id": "id must be a number"},
			},
		})
		return
	}

	err = h.store.DeleteExpenseRule(c.Request.Context(), userID.(int64), id)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "expense rule not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to delete expense rule",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
