package middlewares

import (
	"net/http"

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
        if isRevoked != nil && isRevoked(tokenString) {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
            return
        }
        claims, err := tokens.VerifyToken(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        c.Set("claims", claims)
        c.Next()
    }
}
