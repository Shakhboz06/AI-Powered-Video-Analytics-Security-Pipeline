package api

import (
	"bytes"
	"context"
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

// SignAlertFrameURL asks Supabase storage for a short-lived signed URL of an
// alert's capture frame. Shared by the authenticated and public image routes.
func SignAlertFrameURL(ctx context.Context, alertID string) (string, error) {

	body, err := json.Marshal(gin.H{
		"expiresIn": 300,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%v/storage/v1/object/sign/Alert%%20Frames/frame/%v.jpg", config.GetString("SUPABASE_URL", ""), alertID),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	supabaseKey := config.GetString("SUPABASE_KEY", "")
	httpReq.Header.Set("apikey", supabaseKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call downstream service: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", ErrFrameNotFound
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read downstream response: %w", err)
	}

	var SignedURL struct {
		SignedUrl string `json:"signedURL"`
	}

	if err = json.Unmarshal(bodyBytes, &SignedURL); err != nil {
		return "", fmt.Errorf("failed to parse body: %w", err)
	}

	return fmt.Sprintf("%v/storage/v1%v", config.GetString("SUPABASE_URL", ""), SignedURL.SignedUrl), nil
}

var ErrFrameNotFound = fmt.Errorf("no image found")

func GetAlertImage() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		alertID := ctx.Param("alert_id")

		signedURL, err := SignAlertFrameURL(ctx, alertID)
		if err != nil {
			if err == ErrFrameNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "no image found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"signed_url": signedURL,
		})

	}
}
