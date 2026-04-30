package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"os/signal"
	"syscall"
	"time"
	"video-analytics-pipe/config"
	"video-analytics-pipe/db/postgres"
	myredis "video-analytics-pipe/db/redis/cache"
)

type Detections struct {
	ClassID   int64      `json:"class_id"`
	ConfScore float64    `json:"conf_score"`
	TrackerID int64      `json:"tracker_id"`
	BoundBox  [4]float64 `json:"bound_box"`
	Label     string     `json:"label"`
}

type DetectionResult struct {
	Camera     string       `json:"camera"`
	Detections []Detections `json:"detections"`
	LatencyMS  float64      `json:"latency_ms"`
	RecordedAt int64        `json:"timestamp"`
}

type TabDetection struct {
	Camera          string    `json:"camera"`
	Parameters      []byte    `json:"detections"`
	TotalObjects    []byte    `json:"total_objects"`
	TotalDetections int       `json:"total_detections"`
	LatencyMS       float64   `json:"latency_ms"`
	RecordedAt      time.Time `json:"recorded_at"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	broker := config.GetString("KAFKA_BROKER", "")
	dsn := config.GetString("DB_ADDR", "")

	if dsn == "" {
		log.Fatal("DB_ADDR environment variable is required")
	}

	if broker == "" {
		log.Fatal("KAFKA_BROKER environment variable is required")
	}

	fmt.Printf("Aggregator: Kafka=%s, Database=%s\n",
		broker, dsn,
	)

	specs := []TopicSpec{
		{"video.analysis", 6, 1},
		{"video.results", 6, 1},
	}
	if err := KafkaTopics(ctx, broker, specs); err != nil {
		log.Fatalf("❌  could not create/verify topics: %v", err)
	}

	// ─── Kafka reader from video.results ───────────────────
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "video.results",
		GroupID: "aggregator-group",
	})

	log.Println("▶️  Aggregator listening on video.results …")
	// ─── TimeSeriesDB  ────────────────────────────────────
	database, err := postgres.New()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer reader.Close()

	redisAddr := config.GetString("REDIS_URL", "")
	redisPass := config.GetString("REDIS_PASS", "")
	redisDB := config.GetInt("REDIS_DB", 0)

	if redisAddr == "" {
		log.Fatal("REDIS_URL environment variable is required")
	}

	rdb := myredis.NewRedisClient(redisAddr, redisPass, redisDB)
	defer rdb.Close()

	detStore := NewDetStore(database)
	zoneStore := NewZoneStore(database)
	alertStore := NewAlertStore(database)
	zoneCache := NewZoneCache()
	if err := zoneCache.BackgroundRefresh(ctx, 30*time.Second, zoneStore); err != nil {
		log.Fatal(err)
	}

	intrusionState := NewZoneTrackerState()
	runningDetector := NewRunningDetector()

	for {
		m, err := reader.ReadMessage(ctx)

		if ctx.Err() == context.Canceled {
			log.Printf("Process disrupted or terminated")
			return
		}

		if err != nil {
			log.Printf("Kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var det DetectionResult

		if err := json.Unmarshal(m.Value, &det); err != nil {
			log.Printf("Invalid JSON: %s → %v", string(m.Value), err)
			continue
		}

		var counts = make(map[string]int64)

		for _, item := range det.Detections {
			counts[item.Label] = int64(counts[item.Label] + 1)
		}

		jsonData, err := json.Marshal(det.Detections)
		if err != nil {
			log.Printf("failed to encode to json: %v", err)
			return
		}

		jsonObj, err := json.Marshal(counts)
		if err != nil {
			log.Printf("failed to marshal obj: %v", err)
			return
		}
		var detections TabDetection

		detections = TabDetection{
			Camera:          det.Camera,
			Parameters:      jsonData,
			TotalObjects:    jsonObj,
			TotalDetections: len(det.Detections),
			LatencyMS:       det.LatencyMS,
			RecordedAt:      time.UnixMilli(det.RecordedAt),
		}

		if err := detStore.CreateDetections(ctx, &detections); err != nil {
			log.Printf("failed to insert detection: %v", err)
			return
		}

		if err := detStore.CreateTrackings(ctx, det.Camera, det.RecordedAt, &det.Detections); err != nil {
			log.Printf("failed to insert trackings: %v", err)
			return
		}

		zones := zoneCache.Get(det.Camera)

		var alerts []Alert

		if len(zones) == 0 {
			log.Println("no detected zones for this camera")
		} else {
			currentMap := map[int64]map[int64]Alert{}
			zoneByID := map[int64]Zone{}
			for _, item := range det.Detections {
				x1 := item.BoundBox[0]
				x2 := item.BoundBox[2]
				y2 := item.BoundBox[3]

				point := Point{
					X: (x1 + x2) / 2,
					Y: y2,
				}

				for _, zone := range zones {
					if isZoneActive(zone) {

						if PointinPoligon(point, zone.Polygon) {
							if currentMap[zone.ID] == nil {
								currentMap[zone.ID] = map[int64]Alert{}
							}

							zoneByID[zone.ID] = zone
							currentMap[zone.ID][item.TrackerID] = Alert{
								Camera:     zone.Camera,
								ZoneID:     &zone.ID,
								ZoneName:   zone.Name,
								TrackerID:  item.TrackerID,
								BoundBox:   item.BoundBox,
								Label:      item.Label,
								RecordedAt: detections.RecordedAt,
								Severity:   zone.DefaultSeverity,
							}
						}
					}
				}
			}
			alertIntrusion := intrusionState.UpdateState(currentMap, zoneByID, detections.RecordedAt)
			for _, a := range alertIntrusion {
				alerts = append(alerts, a)
				log.Printf("🚨 NEW INTRUSION: tracker #%d entered zone %d camera %s bound_box %v label %s recorded_at %v alert_type %s", a.TrackerID, a.ZoneID, a.Camera, a.BoundBox, a.Label, a.RecordedAt, a.AlertType)
			}
		}

		runningAlerts := runningDetector.DetectRunning(det.Camera, det.Detections, detections.RecordedAt)
		alerts = append(alerts, runningAlerts...)

		if len(alerts) > 0 {
			if err := alertStore.Create(ctx, alerts); err != nil {
				log.Printf("failed to insert alerts: %v", err)
				return
			}
			publishAlertsToRedis(ctx, rdb, alerts)

		}

		log.Printf("Wrote to TimeSeriesDB: %s lat=%.1fms",
			det.Camera, det.LatencyMS,
		)
	}
}
