package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.FullPath(){
		case "/api/v1/alerts/stream", "/healthz", "/api/v1/public/uploads":
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		// injecting it back into Gin
		c.Request = c.Request.WithContext(ctx)

		c.Next() // continues to handlers
	}
}


