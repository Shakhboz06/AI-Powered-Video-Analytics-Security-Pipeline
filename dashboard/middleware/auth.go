package middleware

import (
	"net/http"
	"strings"
	"video-analytics-pipe/dashboard/internal/auth"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

func AuthByAPIKey(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headerkey := ctx.GetHeader("X-API-Key")
		queryKey := ctx.Query("api_key")
		if headerkey != key && queryKey != key {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "invalid API key"})
			return
		}

		ctx.Next()
	}
}

func AuthTokenMiddleware(s *store.UserStore, auth *auth.JWTAuthenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var tokenStr string

		if cookie, err := ctx.Cookie("auth_token"); err == nil && cookie != ""{
			tokenStr = cookie
		}else{
			authHeader := ctx.GetHeader("Authorization")

			if authHeader == "" {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not Authorized, missing credentials"})
				return
			}
			
			tokenStr = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}
		
		userID, token_ver, err := auth.ValidateToken(tokenStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		version, err := s.GetUserTokenVersion(ctx, userID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve used id"})
			return
		}

		if version.Version != token_ver {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked, log again"})
			return
		}
		ctx.Set("userID", userID)
		ctx.Next()
	}
}
