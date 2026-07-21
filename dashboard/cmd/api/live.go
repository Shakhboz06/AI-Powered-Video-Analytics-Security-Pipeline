package api

import (
	"fmt"
	"io"
	"video-analytics-pipe/dashboard/internal/live"

	"github.com/gin-gonic/gin"
)

func StreamLive(hub *live.Hub) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		camera := ctx.Param("camera")

		ch := hub.Subscribe(camera)
		defer hub.UnSubcribe(camera, ch)

		ctx.Header("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")

		ctx.Stream(func(w io.Writer) bool {

			select {
			case frame, ok := <-ch:
				if !ok {
					return false
				}
				if _, err := fmt.Fprintf(
					w,
					"--frame\r\n"+
						"Content-Type: image/jpeg\r\n"+
						"Content-Length: %d\r\n"+
						"\r\n",
					len(frame),
				); err != nil {
					return false
				}

				if _, err := w.Write(frame); err != nil {
					return false
				}

				if _, err := io.WriteString(w, "\r\n"); err != nil {
					return false
				}
			case <-ctx.Request.Context().Done():
				return false
			}

			return true
		})

	}

}
