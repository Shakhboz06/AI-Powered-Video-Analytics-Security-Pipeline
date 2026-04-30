package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func StreamAlerts(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")

		flusher, ok := ctx.Writer.(http.Flusher)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
			return
		}

		sub := rdb.Subscribe(ctx.Request.Context(), "alerts")
		defer sub.Close()

		ch := sub.Channel()
		
		for {
			select {
			case <-ctx.Request.Context().Done():
				return
			case msg := <-ch:
				fmt.Fprintf(ctx.Writer, "data: %s\n\n", msg.Payload)
				flusher.Flush()
			}
		}
	}
}
