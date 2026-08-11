package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/auth"
	"app/internal/http/dto"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "unauthorized",
					Message: "not authenticated",
				},
			})
			c.Abort()
			return
		}

		claims, err := auth.ParseJWT(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetails{
					Code:    "unauthorized",
					Message: "invalid token",
				},
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
