package api

import (
	"net/http"
	"strconv"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

func GetAllAlerts(s *store.AlertStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		camera := ctx.Query("camera")
		status := ctx.Query("status")

		alerts, err := s.GetAll(ctx, camera, status)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve list of alerts"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"alerts": alerts})
	}
}

func UpdateAlertStatus(s *store.AlertStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var statusPayload struct {
			Status string `json:"status"`
		}

		
		if err := ctx.BindJSON(&statusPayload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
		
		validStatus := map[string]bool{"new": true, "acknowledged": true, "resolved": true}
		if !validStatus[statusPayload.Status] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		strID := ctx.Param("id")

		id, err := strconv.ParseInt(strID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		alert, err := s.UpdateStatus(ctx, id, statusPayload.Status)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update status of alert"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"alert": alert})

	}
}
