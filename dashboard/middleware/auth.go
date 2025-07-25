package middleware

import "github.com/gin-gonic/gin"

func AuthByAPIKey(key string) gin.HandlerFunc{
	return func(ctx *gin.Context){
		headerkey := ctx.GetHeader("X-API-Key")
		if headerkey != key{
			ctx.AbortWithStatusJSON(401, gin.H{"error": "invalid API key"})
			return 
		}

		ctx.Next()
	}
}