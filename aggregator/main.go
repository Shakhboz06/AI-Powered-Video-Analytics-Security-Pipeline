package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"video-analytics-pipe/config"
	"video-analytics-pipe/db/postgres"
	myredis "video-analytics-pipe/db/redis/cache"
)

type Alert struct {
	AlertID    string     `json:"alert_id"`
	Camera     string     `json:"camera"`
	ZoneID     *int64     `json:"zone_id"`
	ZoneName   string     `json:"zone_name,omitempty"`
	TrackerID  int64      `json:"tracker_id"`
	BoundBox   [4]float64 `json:"bound_box"`
	Label      string     `json:"label"`
	RecordedAt time.Time  `json:"recorded_at"`
	AlertType  string     `json:"alert_type"`
	Severity   string     `json:"severity"`
}

type Detections struct {
	ClassID   int64      `json:"class_id"`
	ConfScore float64    `json:"conf_score"`
	TrackerID int64      `json:"tracker_id"`
	BoundBox  [4]float64 `json:"bound_box"`
	Label     string     `json:"label"`
}

type DetectionResult struct {
	Camera            string             `json:"camera"`
	Detections        []Detections       `json:"detections"`
	Keypoints         []Keypoints        `json:"keypoints"`
	FallingDetections map[string]float64 `json:"fall_predictions"`
	FightPredictions  *float64           `json:"fight_predictions"`
	LatencyMS         float64            `json:"latency_ms"`
	RecordedAt        int64              `json:"timestamp"`
}

type TabDetection struct {
	Camera          string    `json:"camera"`
	Parameters      []byte    `json:"detections"`
	TotalObjects    []byte    `json:"total_objects"`
	TotalDetections int       `json:"total_detections"`
	LatencyMS       float64   `json:"latency_ms"`
	RecordedAt      time.Time `json:"recorded_at"`
}

type AlertNotifications struct {
	Camera     string     `json:"camera"`
	AlertID    string     `json:"alert_id"`
	TrackerID  int64      `json:"tracker_id"`
	BoundBox   [4]float64 `json:"bound_box"`
	RecordedAt int64      `json:"timestamp"`
	AlertType  string     `json:"alert_type"`
}

var mlFallCooldowns = make(map[string]map[int64]time.Time)
var mlFallStreaks = make(map[string]map[int64]int)

const mlFallCooldownDuration = 30 * time.Second
const mlConfirmationFrames = 5

func main() {
	shutDownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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
		{"video.alert_notifications", 6, 1},
	}

	if err := KafkaTopics(shutDownCtx, broker, specs); err != nil {
		log.Fatalf("❌  could not create/verify topics: %v", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   "video.alert_notifications",
	})

	defer writer.Close()

	// ─── Kafka reader from video.results ───────────────────
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "video.results",
		GroupID: "aggregator-group",
	})

	log.Println("▶️  Aggregator listening on video.results …")
	defer reader.Close()
	// ─── TimeSeriesDB  ────────────────────────────────────
	database, err := postgres.New()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer database.Close()

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
	keyPointStore := NewKeypointStore(database)

	zoneCache := NewZoneCache()
	if err := zoneCache.BackgroundRefresh(shutDownCtx, 30*time.Second, zoneStore); err != nil {
		log.Fatal(err)
		return
	}

	intrusionState := NewZoneTrackerState()
	runningDetector := NewRunningDetector()
	fallingDetector := NewFallDetector()
	brandishingDetector := NewBrandishingDetector()
	abandonedObjDetector := NewAbandonedObjectDetector()
	fightingDetector := NewFightDetector()

	for {
		select {
		case <-shutDownCtx.Done():
			return
		default:
			m, err := reader.ReadMessage(shutDownCtx)
			if shutDownCtx.Err() == context.Canceled {
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
				continue
			}

			jsonObj, err := json.Marshal(counts)
			if err != nil {
				log.Printf("failed to marshal obj: %v", err)
				continue
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

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := detStore.CreateDetections(ctx, &detections); err != nil {
				log.Printf("failed to insert detection: %v", err)
				continue
			}

			if err := detStore.CreateTrackings(ctx, det.Camera, det.RecordedAt, &det.Detections); err != nil {
				log.Printf("failed to insert trackings: %v", err)
				continue
			}

			if err := keyPointStore.CreateKeyPoints(ctx, det.Camera, det.RecordedAt, det.Keypoints); err != nil {
				log.Printf("failed to insert keypoints: %v", err)
				continue
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

			fallingAlerts := fallingDetector.DetectFall(det.Camera, det.Detections, det.Keypoints, detections.RecordedAt)

			ruleBasedTrackers := map[int64]bool{}
			for _, alert := range fallingAlerts {
				ruleBasedTrackers[alert.TrackerID] = true
			}

			for trackedIDStr, prob := range det.FallingDetections {

				if trackedIDStr == "" {
					continue
				}

				trackedID, err := strconv.ParseInt(trackedIDStr, 10, 64)
				if err != nil {
					log.Printf("failed to parse trackedID: %v", err)
					continue
				}

				over := prob >= 0.7

				log.Printf("falling probability: %v", prob)

				streak := mlRecordFallStreak(det.Camera, trackedID, over)

				if streak < mlConfirmationFrames {
					continue
				}

				if mlCooldownActive(det.Camera, trackedID, detections.RecordedAt) {
					continue
				}

				severity := "medium"

				if ruleBasedTrackers[trackedID] {
					severity = "high"
				}

				var personBbox [4]float64
				bboxFound := false
				for _, item := range det.Detections {
					if item.TrackerID == trackedID && item.Label == "person" {
						personBbox = item.BoundBox
						bboxFound = true
						break
					}
				}

				if !bboxFound {
					continue
				}

				alerts = append(alerts, Alert{
					AlertType:  "falling",
					TrackerID:  trackedID,
					Camera:     det.Camera,
					Severity:   severity,
					RecordedAt: detections.RecordedAt,
					Label:      "person",
					BoundBox:   personBbox,
				})

				updateMLCooldown(det.Camera, trackedID, detections.RecordedAt)
			}

			abandonedObjAlerts := abandonedObjDetector.DetectAbandonedObj(det.Camera, det.Detections, detections.RecordedAt)
			alerts = append(alerts, abandonedObjAlerts...)

			fightingAlerts := fightingDetector.DetectFight(det.Camera, det.Detections, det.FightPredictions, detections.RecordedAt)
			if fightingAlerts != nil {
				alerts = append(alerts, *fightingAlerts)
				log.Printf("fighting alert: %v", fightingAlerts)
			}

			brandishingAlerts := brandishingDetector.DetectBrandishing(det.Camera, det.Detections, det.Keypoints, detections.RecordedAt)
			alerts = append(alerts, brandishingAlerts...)
			if len(alerts) > 0 {
				for i := range alerts {
					alerts[i].AlertID = uuid.New().String()
				}
				if err := alertStore.Create(ctx, alerts); err != nil {
					log.Printf("failed to insert alerts: %v", err)
					continue
				}
				
				for _, a := range alerts{
	
					kafkaAlerts := AlertNotifications{
						Camera: a.Camera,
						AlertID: a.AlertID,
						TrackerID: a.TrackerID,
						RecordedAt: a.RecordedAt.UnixMilli(),
						BoundBox: a.BoundBox,
						AlertType: a.AlertType,
					}
	
					jsonData, err := json.Marshal(kafkaAlerts)
	
					if err != nil{
						log.Printf("failed to encode to json: %v", err)
						continue
					}
	
					err = writer.WriteMessages(shutDownCtx, kafka.Message{
						Key: []byte(a.AlertID),
						Value: jsonData,
					})
					
					if err != nil {
						log.Printf("failed to write the alerts notifications into alerts: %v", err)
						continue
					}
				}
				publishAlertsToRedis(ctx, rdb, alerts)
			}


			cancel()


			log.Printf("Wrote to TimeSeriesDB: %s lat=%.1fms",
				det.Camera, det.LatencyMS,
			)
		}
	}
}

func mlCooldownActive(camera string, trackerID int64, now time.Time) bool {
	if mlFallCooldowns[camera] == nil {
		return false
	}
	lastAlerted, exists := mlFallCooldowns[camera][trackerID]
	if !exists {
		return false
	}
	return now.Sub(lastAlerted) < mlFallCooldownDuration
}

func updateMLCooldown(camera string, trackerID int64, now time.Time) {
	if mlFallCooldowns[camera] == nil {
		mlFallCooldowns[camera] = make(map[int64]time.Time)
	}
	mlFallCooldowns[camera][trackerID] = now
}

func mlRecordFallStreak(camera string, trackerID int64, over bool) int {
	if over {
		if mlFallStreaks[camera] == nil {
			mlFallStreaks[camera] = make(map[int64]int)
		}
		mlFallStreaks[camera][trackerID]++
	} else {
		if mlFallStreaks[camera] != nil {
			mlFallStreaks[camera][trackerID] = 0
		}
	}

	return mlFallStreaks[camera][trackerID]
}
