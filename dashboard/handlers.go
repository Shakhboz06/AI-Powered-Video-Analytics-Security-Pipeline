package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

func getCameras(queryAPI api.QueryAPI, bucket string) gin.HandlerFunc{
	return func(ctx *gin.Context){
		
		measurement := ctx.DefaultQuery("measurement", "frame_detections")
		flux := queryCamera(bucket, measurement)

		res, err := queryAPI.Query(ctx.Request.Context(), flux)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cameras := make([]string, 0)
		for res.Next() {
			val, ok := res.Record().Value().(string)
			if ok {
				cameras = append(cameras, val)
			}
		}
		ctx.JSON(http.StatusOK, gin.H{"cameras": cameras})
	}
}	


func getMetrics(queryAPI api.QueryAPI, bucket string) gin.HandlerFunc{
	return func(ctx *gin.Context) {

		camera := ctx.Query("camera")
		if camera == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "camera parameter required"})
		}

		measurement := ctx.DefaultQuery("measurement", "frame_detections")
		if err := validateParam("measurement", measurement, allowedMeasurements); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		field := ctx.DefaultQuery("field", "cars")
		if err := validateParam("field", field, allowedFields); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		window := ctx.DefaultQuery("window", "1m")

		agg := ctx.DefaultQuery("agg", "max")

		if err := validateParam("agg", agg, allowedAggFns); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		
		// validate field & agg
		flux := queryInflux(bucket, measurement, camera, field, window, agg)

		res, err := queryAPI.Query(ctx.Request.Context(), flux)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if res.Next() {
			ctx.JSON(http.StatusOK, gin.H{
				"camera":       camera,
				"measurement:": measurement,
				"field":        field,
				"value":        res.Record().Value(),
				"time":         res.Record().Time(),
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"message": "no data"})
		}
	}
}



func getMeasurements(queryAPI api.QueryAPI, bucket string) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		
		
		flux := queryMeasurements(bucket)

		res := []string{}
		q, err := queryAPI.Query(ctx.Request.Context(), flux)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal database error"})
		}

		for q.Next() {
			res = append(res, q.Record().Value().(string))
		}
		ctx.JSON(http.StatusOK, gin.H{"measurements": res})
	}
}