package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		// injecting it back into Gin
		c.Request = c.Request.WithContext(ctx)

		c.Next() // continues to handlers
	}
}

// … in main()
// r := gin.Default()
// r.Use(TimeoutMiddleware(2 * time.Second))

// api := r.Group("/api/v1", auth)
// api.GET("/cameras", getCameras(queryAPI, bucket))
// …
