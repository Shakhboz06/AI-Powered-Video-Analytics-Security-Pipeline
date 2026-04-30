package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"video-analytics-pipe/config"
	"video-analytics-pipe/dashboard/internal/auth"
	"video-analytics-pipe/db/redis/cache"
	"video-analytics-pipe/dashboard/middleware"
	"video-analytics-pipe/db/postgres"

	"video-analytics-pipe/dashboard/cmd/api"
	"video-analytics-pipe/dashboard/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// ─── Load config ──────────────────────────────────────────────
	jwtSecret := config.GetString("AUTH_TOKEN_SECRET", "")
	jwtAud := config.GetString("AUTH_JWT_AUD", "video-dashboard")
	jwtIss := config.GetString("AUTH_JWT_ISS", "video-dashboard")
	jwtExp := time.Hour * 2 * 24

	// Create the JWT authenticator
	jwtAuth := auth.NewJWTAuthenticator(jwtSecret, jwtAud, jwtIss, jwtExp)

	// ─── Load configuration ───────────────────────
	addr := config.GetString("DASHBOARD_API_ADDR", "")

	// ────────────   Redis  Caching   ───────────
	redisAddr := config.GetString("REDIS_URL", "")
	redisPass := config.GetString("REDIS_PASS", "")
	redisDB := config.GetInt("REDIS_DB", 0)

	rdb := cache.NewRedisClient(redisAddr, redisPass, redisDB)
	cache.NewRedisStorage(rdb)
	// ─── Router setup ─────────────────────────────
	apiKey := config.GetString("AUTH_API_KEY", "")

	r := gin.Default()

	r.Use(middleware.TimeoutMiddleware(2 * time.Second))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.GetString("CORS_ALLOWED_ORIGIN", config.GetString("FRONTEND_ADDR",""))},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Authorization", "X-CSRF-Token", "X-API-Key"},
		ExposeHeaders:    []string{"Link"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))

	database, err := postgres.New()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	userStore := store.NewUserStore(database)
	r.POST("/api/v1/dashboard/user/register", api.UserRegister(userStore, jwtAuth))
	r.POST("/api/v1/dashboard/user/login", api.UserLogin(userStore, jwtAuth))

	// ✅ Protected routes
	protected := r.Group("/api/v1/dashboard")
	protected.Use(middleware.AuthTokenMiddleware(userStore, jwtAuth))
	protected.GET("/user", api.LoginToDashboard(userStore))

	auth := middleware.AuthByAPIKey(apiKey)
	apis := r.Group("/api/v1", auth)

	detStore := store.NewDetStore(database)
	apis.GET("/detections/cameras", getAllCameras(detStore))
	apis.GET("/detections/metrics", getDetCountsByTime(detStore))
	apis.GET("/detections/live", getDetLiveByCam(detStore))
	apis.GET("/detections/summary", getAggSummary(detStore))
	apis.GET("/detections/class/summary", getClassSummary(detStore))
	apis.GET("/detections/latency", getByLatency(detStore))

	zoneStore := store.NewZonesStore(database)
	apis.POST("/zones", api.CreateZones(zoneStore))
	apis.GET("/zones/list", api.GetAllZones(zoneStore))
	apis.PUT("/zones/:id", api.UpdateZone(zoneStore))
	apis.DELETE("/zones/:id", api.DeleteZone(zoneStore))

	alertStore := store.NewAlertStore(database)
	apis.GET("/alerts", api.GetAllAlerts(alertStore))
	apis.PATCH("/alerts/:id", api.UpdateAlertStatus(alertStore))
	apis.GET("/alerts/stream", api.StreamAlerts(rdb))
	
	cameraStore := store.NewCameraStore(database)
	apis.POST("/cameras", api.CreateCamera(cameraStore))
	apis.GET("cameras", api.GetCameraList(cameraStore))
	apis.PUT("cameras/:id", api.UpdateCamera(cameraStore))
	apis.DELETE("cameras/:id", api.DeleteCamera(cameraStore))

	// ─── Start server ──────────────────────────────
	r.GET("/healthz", func(ctx *gin.Context) {
		if err := database.PingContext(ctx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"check": "healthy"})
		}
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server running on :%v", addr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")

}
