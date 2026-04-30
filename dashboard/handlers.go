package main

import (
	"net/http"
	"strconv"
	"time"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

func getAllCameras(s *store.DetectionStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		camera, err := s.GetAllCamera(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"cameras": camera.Camera})
	}
}

func getDetCountsByTime(s *store.DetectionStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		camera := ctx.Query("camera")

		timerange := ctx.Query("time")

		if camera == "" || timerange == "" {
			ctx.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		t, err := time.Parse(time.RFC3339, timerange)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse time"})
			return
		}

		detections, err := s.GetDetectionCounts(ctx, camera, t)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"detections": []store.DetectionsList{*detections}})
	}
}

func getDetLiveByCam(s *store.DetectionStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		camera := ctx.Query("camera")

		if camera == "" {
			ctx.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		detections, err := s.GetLiveByCam(ctx, camera)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"detections": []store.Detections{*detections}})
	}
}

func getAggSummary(s *store.DetectionStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
		bucket := ctx.Query("bucket")
		camera := ctx.Query("camera")
		st := ctx.Query("start")
		et := ctx.Query("end")

		if bucket == "" || camera == "" || st == "" || et == ""{
			ctx.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		startTime, err := time.Parse(time.RFC3339, st)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse start time"})
			return
		}

		endTime, err := time.Parse(time.RFC3339, et)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse end time"})
			return
		}
		
		summary, err := s.GetAggSummary(ctx, camera, bucket, startTime, endTime)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"aggregated_summary": []store.AggSummary{*summary}})

	}
}

func getClassSummary(s *store.DetectionStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		

		camera := ctx.Query("camera")
		st := ctx.Query("start")
		et := ctx.Query("end")

		if camera == "" || st == "" || et == ""{
			ctx.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		startTime, err := time.Parse(time.RFC3339, st)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse start time"})
			return
		}

		endTime, err := time.Parse(time.RFC3339, et)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse end time"})
			return
		}
		
		summary, err := s.GetClassSum(ctx, camera, startTime, endTime)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}


		ctx.JSON(http.StatusOK, gin.H{"class_summary": []map[string]int64{summary}})

	}
}

func getByLatency(s *store.DetectionStore) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		
		camera := ctx.Query("camera")
		st := ctx.Query("start")
		et := ctx.Query("end")

		if camera == "" || st == "" || et == ""{
			ctx.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		startTime, err := time.Parse(time.RFC3339, st)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse start time"})
			return
		}

		endTime, err := time.Parse(time.RFC3339, et)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to parse end time"})
			return
		}

		threshold := ctx.Query("threshold")
		if threshold == ""{
			threshold = "500"
		}

		f, err := strconv.ParseFloat(threshold, 64)

		if err != nil{
			ctx.JSON(400, gin.H{"error": "failed to convert to float64"})
			return 
		}

		sumLatency, err := s.GetByLatencyMS(ctx, camera, startTime, endTime, f)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"health_summary": sumLatency})

	}
}
