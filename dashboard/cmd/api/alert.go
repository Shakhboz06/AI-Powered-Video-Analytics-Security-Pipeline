package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"video-analytics-pipe/config"
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

func GetAlertImage() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		alertID := ctx.Param("alert_id")

		body, err := json.Marshal(gin.H{
			"expiresIn": 300,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
			return
		}

		httpReq, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			fmt.Sprintf("%v/storage/v1/object/sign/Alert%%20Frames/frame/%v.jpg", config.GetString("SUPABASE_URL", ""), alertID),
			bytes.NewBuffer(body),
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			return
		}
		
		
		supabaseKey := config.GetString("SUPABASE_KEY", "")
		httpReq.Header.Set("apikey", supabaseKey)
		httpReq.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		
		resp, err := client.Do(httpReq)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call downstream service"})
			return
		}
		
		defer resp.Body.Close()
	
		
		if resp.StatusCode != 200 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no image found"})
			return 
		}

		
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to read downstream response",
			})
			return
		}		

		var SignedURL struct {
			SignedUrl string `json:"signedURL"`
		}

		
		if err = json.Unmarshal(bodyBytes, &SignedURL); err != nil{
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to parse body",
			})
			return 
		}

		fullUrl := fmt.Sprintf("%v/storage/v1%v", config.GetString("SUPABASE_URL", ""), SignedURL.SignedUrl)

		ctx.JSON(http.StatusOK, gin.H{
			"signed_url": fullUrl,	
		})

	}
}
