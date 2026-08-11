package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/http/dto"
	"app/internal/store/postgres"
)

type SettingsHandler struct {
	store *postgres.Store
}

func NewSettingsHandler(store *postgres.Store) *SettingsHandler {
	return &SettingsHandler{
		store: store,
	}
}

type UpdateSettingsRequest struct {
	Timezone          string `json:"timezone"`
	AvgWindowDays     int    `json:"avg_window_days"`
	DefaultIncludeAvg bool   `json:"default_include_avg"`
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
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

	var req UpdateSettingsRequest
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

	settings, err := h.store.UpdateUserSettings(
		c.Request.Context(),
		userID.(int64),
		req.Timezone,
		req.AvgWindowDays,
		req.DefaultIncludeAvg,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to update settings",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": gin.H{
			"timezone":            settings.Timezone,
			"avg_window_days":     settings.AvgWindowDays,
			"default_include_avg": settings.DefaultIncludeAvg,
		},
	})
}
