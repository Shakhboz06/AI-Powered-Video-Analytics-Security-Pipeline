package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

type ZonePayload struct {
	Name            string        `json:"name,omitempty"`
	Camera          string        `json:"camera,omitempty"`
	Polygon         []store.Point `json:"polygon,omitempty"`
	IsActive        bool          `json:"is_active"`
	ActiveFrom      *time.Time    `json:"active_from,omitempty"`
	ActiveUntil     *time.Time    `json:"active_until,omitempty"`
	LoiterThreshold *int          `json:"loiter_threshold_seconds"`
	DefaultSeverity string        `json:"default_severity"`
}

func CreateZones(s *store.ZoneStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var payload ZonePayload
		payload.IsActive = true
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		zone := &store.Zone{
			Name:            payload.Name,
			Camera:          payload.Camera,
			Polygon:         payload.Polygon,
			IsActive:        payload.IsActive,
			ActiveFrom:      payload.ActiveFrom,
			ActiveUntil:     payload.ActiveUntil,
			LoiterThreshold: payload.LoiterThreshold,
			DefaultSeverity: payload.DefaultSeverity,
		}

		zones, err := s.Create(ctx, zone)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not save zone"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"data": zones,
		})
	}
}

func GetAllZones(s *store.ZoneStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		camera := ctx.Query("camera")

		if camera == "" {
			ctx.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		zones, err := s.Get(ctx, camera)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve list of zones"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": zones,
		})
	}
}

func UpdateZone(s *store.ZoneStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var payload ZonePayload
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		zone := store.Zone{
			Name:            payload.Name,
			Camera:          payload.Camera,
			Polygon:         payload.Polygon,
			IsActive:        payload.IsActive,
			ActiveFrom:      payload.ActiveFrom,
			ActiveUntil:     payload.ActiveUntil,
			LoiterThreshold: payload.LoiterThreshold,
			DefaultSeverity: payload.DefaultSeverity,
		}

		id := ctx.Param("id")

		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		updZone, err := s.Update(ctx, idInt, &zone)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update zone"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": updZone,
		})

	}
}

func DeleteZone(s *store.ZoneStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := s.Delete(ctx, idInt); err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete zone"})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
