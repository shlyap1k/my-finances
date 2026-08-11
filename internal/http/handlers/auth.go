package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/auth"
	"app/internal/http/dto"
	"app/internal/store/postgres"
)

type AuthHandler struct {
	store     *postgres.Store
	jwtSecret string
}

func NewAuthHandler(store *postgres.Store, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
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

	// Check if user already exists
	existingUser, _ := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "conflict",
				Message: "email already registered",
				Fields:  map[string]string{"email": "email is already registered"},
			},
		})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to hash password",
			},
		})
		return
	}

	// Create user
	user, err := h.store.CreateUser(c.Request.Context(), req.Email, passwordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to create user",
			},
		})
		return
	}

	// Create default user settings
	_, err = h.store.CreateUserSettings(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to create user settings",
			},
		})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to generate token",
			},
		})
		return
	}

	// Set cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true)

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
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

	// Get user by email
	user, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "invalid credentials",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get user",
			},
		})
		return
	}

	// Check password
	if err := auth.CheckPasswordHash(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "unauthorized",
				Message: "invalid credentials",
			},
		})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to generate token",
			},
		})
		return
	}

	// Set cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the cookie by setting it with a past expiration
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, dto.SuccessResponse{Status: "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
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

	user, err := h.store.GetUserByID(c.Request.Context(), userID.(int64))
	if err == postgres.ErrNotFound {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "not_found",
				Message: "user not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get user",
			},
		})
		return
	}

	settings, err := h.store.GetUserSettings(c.Request.Context(), user.ID)
	if err == postgres.ErrNotFound {
		// Create default settings if they don't exist
		settings, err = h.store.CreateUserSettings(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "internal_error",
					Message: "failed to create user settings",
				},
			})
			return
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetails{
				Code:    "internal_error",
				Message: "failed to get user settings",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
		"settings": gin.H{
			"timezone":            settings.Timezone,
			"avg_window_days":     settings.AvgWindowDays,
			"default_include_avg": settings.DefaultIncludeAvg,
		},
	})
}
