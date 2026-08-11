package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"app/internal/http/dto"
	"app/internal/store/postgres"
)

type SavingsRuleHandler struct {
	store *postgres.Store
}

func NewSavingsRuleHandler(store *postgres.Store) *SavingsRuleHandler {
	return &SavingsRuleHandler{
		store: store,
	}
}

type CreateSavingsRuleRequest struct {
	IncomeRuleID       *int64 `json:"income_rule_id,omitempty"`
	SavingsAccountID   int64  `json:"savings_account_id" binding:"required"`
	RuleType           string `json:"rule_type" binding:"required,oneof=percent fixed"`
	PercentBps         *int32 `json:"percent_bps,omitempty"`
	FixedAmountMinor   *int64 `json:"fixed_amount_minor,omitempty"`
	Priority           int    `json:"priority" binding:"required,min=1"`
}

type UpdateSavingsRuleRequest struct {
	IncomeRuleID       *int64 `json:"income_rule_id,omitempty"`
	SavingsAccountID   int64  `json:"savings_account_id" binding:"required"`
	RuleType           string `json:"rule_type" binding:"required,oneof=percent fixed"`
	PercentBps         *int32 `json:"percent_bps,omitempty"`
	FixedAmountMinor   *int64 `json:"fixed_amount_minor,omitempty"`
	Priority           int    `json:"priority" binding:"required,min=1"`
}

func (h *SavingsRuleHandler) ListSavingsRules(c *gin.Context) {
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

	rules, err := h.store.GetSavingsRules(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get savings rules",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": rules,
	})
}

func (h *SavingsRuleHandler) GetSavingsRule(c *gin.Context) {
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

	rule, err := h.store.GetSavingsRuleByID(c.Request.Context(), userID.(int64), id)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "savings rule not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get savings rule",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"savings_rule": rule,
	})
}

func (h *SavingsRuleHandler) CreateSavingsRule(c *gin.Context) {
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

	var req CreateSavingsRuleRequest
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

	// Валидация типа правила и соответствующих полей
	if req.RuleType == "percent" {
		if req.PercentBps == nil || *req.PercentBps < 0 || *req.PercentBps > 10000 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "validation",
					Message: "invalid request",
					Fields:  map[string]string{"percent_bps": "must be between 0 and 10000 for percent type"},
				},
			})
			return
		}
	} else if req.RuleType == "fixed" {
		if req.FixedAmountMinor == nil || *req.FixedAmountMinor <= 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "validation",
					Message: "invalid request",
					Fields:  map[string]string{"fixed_amount_minor": "must be greater than 0 for fixed type"},
				},
			})
			return
		}
	}

	rule, err := h.store.CreateSavingsRule(
		c.Request.Context(),
		userID.(int64),
		req.IncomeRuleID,
		req.SavingsAccountID,
		req.RuleType,
		req.PercentBps,
		req.FixedAmountMinor,
		req.Priority,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to create savings rule",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"savings_rule": rule,
	})
}

func (h *SavingsRuleHandler) UpdateSavingsRule(c *gin.Context) {
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

	var req UpdateSavingsRuleRequest
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

	// Валидация типа правила и соответствующих полей
	if req.RuleType == "percent" {
		if req.PercentBps == nil || *req.PercentBps < 0 || *req.PercentBps > 10000 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "validation",
					Message: "invalid request",
					Fields:  map[string]string{"percent_bps": "must be between 0 and 10000 for percent type"},
				},
			})
			return
		}
	} else if req.RuleType == "fixed" {
		if req.FixedAmountMinor == nil || *req.FixedAmountMinor <= 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "validation",
					Message: "invalid request",
					Fields:  map[string]string{"fixed_amount_minor": "must be greater than 0 for fixed type"},
				},
			})
			return
		}
	}

	rule, err := h.store.UpdateSavingsRule(
		c.Request.Context(),
		userID.(int64),
		id,
		req.IncomeRuleID,
		req.SavingsAccountID,
		req.RuleType,
		req.PercentBps,
		req.FixedAmountMinor,
		req.Priority,
	)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "savings rule not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to update savings rule",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"savings_rule": rule,
	})
}

func (h *SavingsRuleHandler) DeleteSavingsRule(c *gin.Context) {
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

	err = h.store.DeleteSavingsRule(c.Request.Context(), userID.(int64), id)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "savings rule not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to delete savings rule",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
