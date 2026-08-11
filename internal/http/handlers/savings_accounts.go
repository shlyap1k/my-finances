package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"app/internal/http/dto"
	"app/internal/store/postgres"
)

type SavingsAccountHandler struct {
	store *postgres.Store
}

func NewSavingsAccountHandler(store *postgres.Store) *SavingsAccountHandler {
	return &SavingsAccountHandler{
		store: store,
	}
}

type CreateSavingsAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateSavingsAccountRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *SavingsAccountHandler) ListSavingsAccounts(c *gin.Context) {
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

	accounts, err := h.store.GetSavingsAccounts(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get savings accounts",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": accounts,
	})
}

func (h *SavingsAccountHandler) GetSavingsAccount(c *gin.Context) {
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

	account, err := h.store.GetSavingsAccountByID(c.Request.Context(), userID.(int64), id)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "savings account not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get savings account",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"savings_account": account,
	})
}

func (h *SavingsAccountHandler) CreateSavingsAccount(c *gin.Context) {
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

	var req CreateSavingsAccountRequest
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

	account, err := h.store.CreateSavingsAccount(
		c.Request.Context(),
		userID.(int64),
		req.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to create savings account",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"savings_account": account,
	})
}

func (h *SavingsAccountHandler) UpdateSavingsAccount(c *gin.Context) {
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

	var req UpdateSavingsAccountRequest
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

	account, err := h.store.UpdateSavingsAccount(
		c.Request.Context(),
		userID.(int64),
		id,
		req.Name,
	)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "savings account not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to update savings account",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"savings_account": account,
	})
}

func (h *SavingsAccountHandler) DeleteSavingsAccount(c *gin.Context) {
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

	err = h.store.DeleteSavingsAccount(c.Request.Context(), userID.(int64), id)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "savings account not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to delete savings account",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
