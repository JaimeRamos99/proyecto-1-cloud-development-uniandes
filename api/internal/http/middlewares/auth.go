package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"proyecto1/root/internal/auth"
)

// AuthMiddleware verifies JWT and optionally checks server-side revocation.
// Pass a function to check whether a token has been revoked (e.g., in-memory or Redis),
// or nil if you only want stateless JWT validation.
func AuthMiddleware(tokens auth.TokenManager, isRevoked func(token string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenString := authHeader[len(prefix):]

		// Check revocation
		if isRevoked != nil && isRevoked(tokenString) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		// Verify and parse claims
		claims, err := tokens.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Extract user_id from claims
		if rawID, ok := claims["user_id"]; ok {
			switch v := rawID.(type) {
			case string:
				if id, err := strconv.Atoi(v); err == nil {
					c.Set("userID", id)
				}
			case float64:
				c.Set("userID", int(v))
			default:
				// unsupported type
			}
		}

		// Still set all claims if you want them
		c.Set("claims", claims)

		c.Next()
	}
}
