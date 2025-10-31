package main

import (
	"net/http"
	"time"

	"video-analytics-pipe/config"
	"video-analytics-pipe/dashboard/internal/auth"
	"video-analytics-pipe/dashboard/internal/db"
	"video-analytics-pipe/dashboard/middleware"

	"video-analytics-pipe/dashboard/cmd/api"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {

	// ─── Load config ──────────────────────────────────────────────
	jwtSecret := config.GetString("AUTH_TOKEN_SECRET", "")
	jwtAud := config.GetString("AUTH_JWT_AUD", "video-dashboard")
	jwtIss := config.GetString("AUTH_JWT_ISS", "video-dashboard")
	jwtExp := time.Hour * 2* 24
	
	// Create the JWT authenticator
	jwtAuth := auth.NewJWTAuthenticator(jwtSecret, jwtAud, jwtIss, jwtExp)

	// ─── Load configuration ───────────────────────
	influxURL := config.GetString("INFLUX_URL", "http://influxdb:8086")
	token := config.GetString("INFLUX_TOKEN", "")
	org := config.GetString("INFLUX_ORG", "example_org")
	bucket := config.GetString("INFLUX_BUCKET", "video_metrics")
	addr := config.GetString("DASHBOARD_API_ADDR", ":8081")
	// ─── InfluxDB client & Query API ─────────────
	client := influxdb2.NewClient(influxURL, token)
	queryAPI := client.QueryAPI(org)
	defer client.Close()

	// ─── Router setup ─────────────────────────────
	apiKey := config.GetString("AUTH_API_KEY", "")

	r := gin.Default()

	r.Use(middleware.TimeoutMiddleware(2 * time.Second))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.GetString("CORS_ALLOWED_ORIGIN", "http://*")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Content-Type", "Authorization", "X-CSRF-Token", "X-API-Key"},
		ExposeHeaders:    []string{"Link"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		if _, err := client.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"check": "healthy"})
		}
	})

	database, err := db.New()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	store := store.NewUserStore(database)
	r.POST("/api/v1/dashboard/user/register", api.UserRegister(store, jwtAuth))
	r.POST("/api/v1/dashboard/user/login", api.UserLogin(store, jwtAuth))
		
	// ✅ Protected routes
	protected := r.Group("/api/v1/dashboard")
	protected.Use(middleware.AuthTokenMiddleware(store, jwtAuth))
	protected.GET("/user", api.LoginToDashboard(store))

	auth := middleware.AuthByAPIKey(apiKey)
	api := r.Group("/api/v1", auth)

	api.GET("/measurements", getMeasurements(queryAPI, bucket))

	// GET /api/v1/cameras?measurement={frame_detections|video_enriched}
	api.GET("/cameras", getCameras(queryAPI, bucket))

	// GET /api/v1/metrics?camera=&field=&window=&agg=
	api.GET("/metrics", getMetrics(queryAPI, bucket))
	// ─── Start server ──────────────────────────────
	r.Run(addr)
}
