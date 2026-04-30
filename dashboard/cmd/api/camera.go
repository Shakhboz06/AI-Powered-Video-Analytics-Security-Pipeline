package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

type CameraPayload struct {
	ID          int64     `json:"camera_id"`
	Name        string    `json:"camera_name"`
	VideoSource string    `json:"video_source"`
	IsActive    bool      `json:"is_active"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func CreateCamera(s *store.CameraStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var payload CameraPayload
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		camera := &store.Cameras{
			Name:        payload.Name,
			VideoSource: payload.VideoSource,
			IsActive:    payload.IsActive,
		}

		cameras, err := s.Create(ctx, camera)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not save camera"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"camera": cameras})

	}
}

func GetCameraList(s *store.CameraStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		cameraList, err := s.GetAll(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrive list of camera"})
			return 
		}

		ctx.JSON(http.StatusOK, gin.H{"cameras": cameraList})
	}
}

func UpdateCamera(s *store.CameraStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var payload CameraPayload
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		camera := store.Cameras{
			Name:        payload.Name,
			VideoSource: payload.VideoSource,
			IsActive:    payload.IsActive,
		}

		camIdString := ctx.Param("id")
		id, err := strconv.ParseInt(camIdString, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		cameras, err := s.Update(ctx, camera, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update camera"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"camera": cameras})
	}
}

func DeleteCamera(s *store.CameraStore) gin.HandlerFunc{
	return func(ctx *gin.Context) {

		camIdString := ctx.Param("id")
		id, err := strconv.ParseInt(camIdString, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := s.Delete(ctx, id); err != nil{
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "camera not found"})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete camera"})
			return 
		}

		ctx.Status(http.StatusNoContent)
	}
}