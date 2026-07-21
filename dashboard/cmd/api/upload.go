package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"video-analytics-pipe/config"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var allowedVideoExts = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".avi":  true,
	".mkv":  true,
	".webm": true,
}


type UploadJobMessage struct {
	JobID            string `json:"job_id"`
	StreamID         string `json:"stream_id"`
	Path             string `json:"path"`
	OriginalFilename string `json:"original_filename"`
}


func CreateUpload(jobs *store.JobStore, writer *kafka.Writer, uploadDir string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		maxBytes := int64(config.GetInt("UPLOAD_MAX_BYTES", 200*1024*1024))
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)

		file, header, err := ctx.Request.FormFile("video")
		if err != nil {
			log.Println("upload create error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing or oversized 'video' file (max %d MB)", maxBytes/(1024*1024))})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if !allowedVideoExts[ext] {
			log.Println("upload create error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type, expected one of: mp4, mov, avi, mkv, webm"})
			return
		}

		jobID := uuid.New().String()
		streamID := "upload-" + jobID

		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			log.Println("upload create error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not prepare upload storage"})
			return
		}

		dstPath := filepath.Join(uploadDir, jobID+ext)
		dst, err := os.Create(dstPath)
		if err != nil {
			log.Println("upload create error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not store uploaded file"})
			return
		}

		if _, err := io.Copy(dst, file); err != nil {
			log.Println("upload create error", err)
			dst.Close()
			os.Remove(dstPath)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("upload interrupted or exceeds %d MB", maxBytes/(1024*1024))})
			return
		}
		dst.Close()

		job := &store.AnalysisJob{
			JobID:            jobID,
			StreamID:         streamID,
			OriginalFilename: filepath.Base(header.Filename),
			Status:           "queued",
		}

		if err := jobs.Create(ctx, job); err != nil {
			log.Println("upload create error", err)
			os.Remove(dstPath)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create analysis job"})
			return
		}

		msg := UploadJobMessage{
			JobID:            jobID,
			StreamID:         streamID,
			Path:             dstPath,
			OriginalFilename: job.OriginalFilename,
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			log.Println("upload create error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not encode analysis job"})
			return
		}

		writeCtx, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		if err := writer.WriteMessages(writeCtx, kafka.Message{
			Key:   []byte(streamID),
			Value: payload,
		}); err != nil {
			log.Println("upload create error", err)
			errMsg := "analysis queue unavailable"
			_ = jobs.UpdateProgress(context.Background(), jobID, "failed", 0, &errMsg, nil)
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "analysis queue unavailable, try again later"})
			return
		}

		ctx.JSON(http.StatusAccepted, gin.H{"job": job})
	}
}


func GetUploadJob(jobs *store.JobStore, alerts *store.AlertStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jobID, ok := parseJobID(ctx)
		if !ok {
			return
		}

		job, err := jobs.GetByJobID(ctx, jobID)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not load analysis"})
			return
		}

		if err := jobs.FinalizeIfComplete(ctx, job); err != nil {
			// non-fatal: keep reporting "finalizing" and let the next poll retry
			fmt.Printf("finalize check failed for job %s: %v\n", jobID, err)
		}

		jobAlerts, err := alerts.GetForStream(ctx, job.StreamID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not load alerts"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"job":    job,
			"alerts": jobAlerts,
		})
	}
}


func GetUploadAlertImage(jobs *store.JobStore, alerts *store.AlertStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jobID, ok := parseJobID(ctx)
		if !ok {
			return
		}

		alertID := ctx.Param("alert_id")
		if _, err := uuid.Parse(alertID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert id"})
			return
		}

		job, err := jobs.GetByJobID(ctx, jobID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
			return
		}

		owned, err := alerts.AlertBelongsToStream(ctx, alertID, job.StreamID)
		if err != nil || !owned {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}

		signedURL, err := SignAlertFrameURL(ctx, alertID)
		if err != nil {
			if err == ErrFrameNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "no image found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign image"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"signed_url": signedURL})
	}
}


func UpdateUploadJob(jobs *store.JobStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jobID, ok := parseJobID(ctx)
		if !ok {
			return
		}

		var payload struct {
			Status      string  `json:"status"`
			Progress    int     `json:"progress"`
			Error       *string `json:"error"`
			LastFrameMS *int64  `json:"last_frame_ms"`
		}

		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		validStatus := map[string]bool{"processing": true, "finalizing": true, "failed": true}
		if !validStatus[payload.Status] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}

		if payload.Progress < 0 {
			payload.Progress = 0
		}
		if payload.Progress > 100 {
			payload.Progress = 100
		}

		var lastFrameAt *time.Time
		if payload.LastFrameMS != nil {
			t := time.UnixMilli(*payload.LastFrameMS)
			lastFrameAt = &t
		}

		if err := jobs.UpdateProgress(ctx, jobID, payload.Status, payload.Progress, payload.Error, lastFrameAt); err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update analysis"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func parseJobID(ctx *gin.Context) (string, bool) {
	jobID := ctx.Param("job_id")
	if _, err := uuid.Parse(jobID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return "", false
	}
	return jobID, true
}

