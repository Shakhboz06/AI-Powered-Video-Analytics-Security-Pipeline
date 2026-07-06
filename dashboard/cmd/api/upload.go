package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"video-analytics-pipe/config"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// UploadJobMessage is published on video.upload_jobs for the upload ingestor.
type UploadJobMessage struct {
	JobID    string `json:"job_id"`
	VideoURL string `json:"video_url"`
}

var allowedUploadExts = map[string]bool{
	".mp4": true,
	".avi": true,
	".mov": true,
}

func uploadsBucket() string {
	// Separate bucket from evidence frames so retention policies stay
	// independent (uploads get deleted after processing, evidence never).
	return config.GetString("SUPABASE_UPLOADS_BUCKET", "uploads")
}

func uploadObjectPath(jobID string) string {
	return fmt.Sprintf("uploads/%s.mp4", jobID)
}

// uploadVideoToStorage streams the file body to Supabase storage — the body
// is passed as a reader so large files are never held in memory.
func uploadVideoToStorage(ctx context.Context, objectPath string, body io.Reader) error {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%v/storage/v1/object/%v/%v", config.GetString("SUPABASE_URL", ""), uploadsBucket(), objectPath),
		body,
	)
	if err != nil {
		return fmt.Errorf("failed to create storage request: %w", err)
	}

	supabaseKey := config.GetString("SUPABASE_KEY", "")
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Content-Type", "video/mp4")

	// Big files over residential upload are slow — much longer than the sign call.
	client := &http.Client{Timeout: 120 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload to storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("storage upload returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// signUploadURL mints a signed URL for the uploaded video, long enough to
// cover processing (~1 hour). Same REST pattern as the alert-frame sign call.
func signUploadURL(ctx context.Context, objectPath string) (string, error) {

	payload, err := json.Marshal(gin.H{"expiresIn": 3600})
	if err != nil {
		return "", fmt.Errorf("failed to marshal sign request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%v/storage/v1/object/sign/%v/%v", config.GetString("SUPABASE_URL", ""), uploadsBucket(), objectPath),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create sign request: %w", err)
	}

	supabaseKey := config.GetString("SUPABASE_KEY", "")
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call sign endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("sign endpoint returned %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read sign response: %w", err)
	}

	var signed struct {
		SignedUrl string `json:"signedURL"`
	}
	if err := json.Unmarshal(respBody, &signed); err != nil {
		return "", fmt.Errorf("failed to parse sign response: %w", err)
	}

	return fmt.Sprintf("%v/storage/v1%v", config.GetString("SUPABASE_URL", ""), signed.SignedUrl), nil
}

// deleteUploadObject removes the source video after processing — evidence
// frames and alerts persist, the footage itself doesn't (storage limitation).
func deleteUploadObject(ctx context.Context, objectPath string) error {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%v/storage/v1/object/%v/%v", config.GetString("SUPABASE_URL", ""), uploadsBucket(), objectPath),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	supabaseKey := config.GetString("SUPABASE_KEY", "")
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call delete endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("delete endpoint returned %d", resp.StatusCode)
	}

	return nil
}

// CreateUpload handles POST /api/uploads: validates the video, stores it in
// Supabase, records the job and enqueues it on video.upload_jobs.
func CreateUpload(uploads *store.UploadStore, writer *kafka.Writer) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		maxBytes := int64(config.GetInt("UPLOAD_MAX_BYTES", 200*1024*1024))
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)

		file, header, err := ctx.Request.FormFile("video")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing or oversized 'video' file (max %d MB)", maxBytes/(1024*1024))})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if !allowedUploadExts[ext] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type, expected mp4, avi or mov"})
			return
		}

		if header.Size > maxBytes {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file exceeds the %d MB limit", maxBytes/(1024*1024))})
			return
		}

		jobID := uuid.New().String()
		objectPath := uploadObjectPath(jobID)

		if err := uploadVideoToStorage(ctx.Request.Context(), objectPath, file); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not store the uploaded video"})
			return
		}

		if err := uploads.Create(ctx, jobID, filepath.Base(header.Filename)); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create upload job"})
			return
		}

		videoURL, err := signUploadURL(ctx.Request.Context(), objectPath)
		if err != nil {
			_ = uploads.UpdateStatus(ctx, jobID, "failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not prepare the video for analysis"})
			return
		}

		payload, err := json.Marshal(UploadJobMessage{JobID: jobID, VideoURL: videoURL})
		if err != nil {
			_ = uploads.UpdateStatus(ctx, jobID, "failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not encode upload job"})
			return
		}

		writeCtx, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		// DB row first, then the Kafka message; a failed publish marks the row failed.
		if err := writer.WriteMessages(writeCtx, kafka.Message{
			Key:   []byte(jobID),
			Value: payload,
		}); err != nil {
			_ = uploads.UpdateStatus(ctx, jobID, "failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "analysis queue unavailable, try again later"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"job_id": jobID})
	}
}

// GetUpload handles GET /api/uploads/:job_id — the status endpoint polled by
// the results page and used by the ingestor's idempotency check.
func GetUpload(uploads *store.UploadStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jobID := ctx.Param("job_id")
		if _, err := uuid.Parse(jobID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
			return
		}

		job, err := uploads.Get(ctx, jobID)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not load upload"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"job": job})
	}
}

// UpdateUploadStatus handles PATCH /api/uploads/:job_id/status — the internal
// write path the upload ingestor uses (API-key protected). Terminal states
// also delete the source video from storage: it served its purpose.
func UpdateUploadStatus(uploads *store.UploadStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jobID := ctx.Param("job_id")
		if _, err := uuid.Parse(jobID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
			return
		}

		var payload struct {
			Status string `json:"status"`
		}
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		validStatus := map[string]bool{"processing": true, "done": true, "failed": true}
		if !validStatus[payload.Status] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}

		if err := uploads.UpdateStatus(ctx, jobID, payload.Status); err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update upload"})
			return
		}

		if payload.Status == "done" || payload.Status == "failed" {
			// Best-effort, own context: the 2s request timeout must not
			// leave the source video behind.
			deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := deleteUploadObject(deleteCtx, uploadObjectPath(jobID)); err != nil {
				fmt.Printf("could not delete uploaded video for job %s: %v\n", jobID, err)
			}
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
