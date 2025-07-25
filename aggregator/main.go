// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net"
// 	"time"
// 	"video-analytics-pipe/config"

// 	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
// 	// "github.com/joho/godotenv"
// 	"github.com/segmentio/kafka-go"
// )

// type DetectionResult struct {
// 	Camera    string  `json:"camera"`
// 	Persons   int  `json:"persons"`
// 	Cars      int  `json:"cars"`
// 	LatencyMS float64 `json:"latency_ms"`
// 	TimeStamp int64 `json:"timestamp"`
// }

// type EnrichedEvent struct {
//     DetectionResult
//     AvgLatency1m            float64 `json:"avg_latency_1m"`
//     CarCountThresholdExceeded bool   `json:"high_traffic"`
// }


// func retryKafka(broker string, timeout time.Duration) {
// 	deadline := time.Now().Add(timeout)
// 	for time.Now().Before(deadline) {
// 		conn, err := net.DialTimeout("tcp", broker, 1*time.Second)
// 		if err == nil {
// 			conn.Close()
// 			log.Printf("kafka is up at %s", broker)
// 			return
// 		}
// 		log.Printf("Waiting for Kafka at %s… (%v)", broker, err)
// 		time.Sleep(1 * time.Second)
// 	}
// 	log.Fatalf("timeout waiting for Kafka at %s", broker)
// }
// func main() {

// 	// if err := godotenv.Load("../.env"); err != nil {
// 	//     log.Printf("could not load .env: %v", err)
// 	// }

	
// 	influxURL := config.GetString("INFLUX_URL", "http://localhost:8086")
// 	token := config.GetString("INFLUX_TOKEN", "")
// 	org := config.GetString("INFLUX_ORG", "")
// 	bucket := config.GetString("INFLUX_BUCKET", "")

// 	fmt.Printf("Connecting to %s; org=%s, bucket=%s\n", influxURL, org, bucket)

// 	//InfluxDB client & APIs
// 	client := influxdb2.NewClient(influxURL, token)
// 	defer client.Close()
// 	writeAPI := client.WriteAPIBlocking(org, bucket)
// 	quaryAPI := client.QueryAPI(org)

// 	broker := config.GetString("KAFKA_BROKER", "kafka:9092")
// 	topic := config.GetString("KAFKA_TOPIC", "video.results")
// 	enrichedTopic := config.GetString("KAFKA_ENRICHED_TOPIC", "video.enriched")

// 	retryKafka(broker, 60*time.Second)

// 	specs := []TopicSpec{
// 		{Name: "video.analysis", NumPartitions: 3, ReplicationFactor: 1},
// 		{Name: "video.results",  NumPartitions: 3, ReplicationFactor: 1},
// 		{Name: enrichedTopic,     NumPartitions: 3, ReplicationFactor: 1},
// 	}

// 	if err := KafkaTopics(context.Background(), broker, specs); err != nil {
// 		log.Fatalf("⚠️  failed ensuring topics: %v", err)
// 	}



// 	reader := kafka.NewReader(kafka.ReaderConfig{
// 		Brokers: []string{broker},
// 		Topic:   topic,
// 		GroupID: "aggregator-group",
// 	})
// 	defer reader.Close()


// 	// Writer for enriched events
// 	enricher := kafka.NewWriter(kafka.WriterConfig{
// 		Brokers: []string{broker},
// 		Topic:   enrichedTopic,
// 	})
// 	defer enricher.Close()

// 	log.Println("▶️  Aggregator up, reading from", topic, "and producing to", enrichedTopic)
	
// 	for {
// 		mes, err := reader.ReadMessage(context.Background())
// 		if err != nil {
// 			fmt.Printf("Error reading message: %v\n", err)
// 			time.Sleep(time.Second)
// 			continue
// 		}

// 		var detection DetectionResult
// 		if err := json.Unmarshal(mes.Value, &detection); err != nil {
// 			log.Printf("JSON unmarshal error: %v", err)
// 			continue
// 		}

		
// 		// 3) Write raw point into InfluxDB
// 		p := influxdb2.NewPoint(
// 			"frame_detections",
// 			map[string]string{"camera": detection.Camera},
// 			map[string]interface{}{
// 				"persons": detection.Persons, 
// 				"cars": detection.Cars,
// 				"latency_ms": detection.LatencyMS,
// 			},
// 			time.UnixMilli(detection.TimeStamp),
// 		)

// 		if err := writeAPI.WritePoint(context.Background(), p); err != nil {
// 			fmt.Printf("Failed to write: %v\n", err)
// 			continue
// 		}

// 		log.Printf("Wrote to InfluxDB: %s persons=%d cars=%d lat=%.1fms", detection.Camera, detection.Persons, detection.Cars, detection.LatencyMS)

// 		// 4) Query rolling aggregates (last 1 minute)
// 		avgLat, err := queryInflux(quaryAPI, bucket, detection.Camera, "latency_ms", "mean", 1)
// 		if err != nil {
// 			log.Printf("Influx avg query error: %v", err)
// 		}

// 		maxCars, err := queryInflux(quaryAPI, bucket, detection.Camera, "cars", "max", 1)
// 		if err != nil {
// 			log.Printf("❌ Influx max query error: %v", err)
// 		}


// 		// 5) Build enriched event
// 		res := EnrichedEvent{
// 			DetectionResult:           detection,
// 			AvgLatency1m:              avgLat,
// 			CarCountThresholdExceeded: maxCars > 10,
// 		}



// 		out, err := json.Marshal(res)
// 		if err != nil {
// 			log.Printf("JSON marshalling error: %v", err)
// 			continue
// 		}

// 		err = enricher.WriteMessages(context.Background(),kafka.Message{Value: out},)
// 		if err != nil {
// 			log.Printf("Enriched write error: %v", err)
// 		}
// 	}

// }


// aggregator/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
    "video-analytics-pipe/config"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/segmentio/kafka-go"
)

type DetectionResult struct {
    Camera    string  `json:"camera"`
    Persons   int     `json:"persons"`
    Cars      int     `json:"cars"`
    LatencyMS float64 `json:"latency_ms"`
    Timestamp int64   `json:"timestamp"`
}


func main() {
    ctx := context.Background()

    // ─── Load config from .env ──────────────────────────────
    broker := config.GetString("KAFKA_BROKER", "kafka:9092")
    influxURL := config.GetString("INFLUX_URL", "http://localhost:8086")
    token     := config.GetString("INFLUX_TOKEN", "")
    org       := config.GetString("INFLUX_ORG", "")
    bucket    := config.GetString("INFLUX_BUCKET", "")

    fmt.Printf("Aggregator: Kafka=%s, InfluxDB=%s (org=%s, bucket=%s)\n",
        broker, influxURL, org, bucket,
    )

    // ─── Ensure the two topics exist ───────────────────────
    specs := []TopicSpec{
        {"video.analysis", 1, 1},
        {"video.results",  1, 1},
    }
    if err := KafkaTopics(ctx, broker, specs); err != nil {
        log.Fatalf("❌  could not create/verify topics: %v", err)
    }

    // ─── InfluxDB client ────────────────────────────────────
    client   := influxdb2.NewClient(influxURL, token)
    writeAPI := client.WriteAPIBlocking(org, bucket)
    defer client.Close()

    // ─── Kafka reader from video.results ───────────────────
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{broker},
        Topic:   "video.results",      // ← this must match what the worker produces
        GroupID: "aggregator-group",
    })
    defer reader.Close()

    log.Println("▶️  Aggregator listening on video.results …")

    for {
        m, err := reader.ReadMessage(ctx)
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

        // Write to InfluxDB
        point := influxdb2.NewPoint(
            "frame_detections",
            map[string]string{"camera": det.Camera},
            map[string]interface{}{
                "persons":    det.Persons,
                "cars":       det.Cars,
                "latency_ms": det.LatencyMS,
            },
            time.UnixMilli(det.Timestamp),
        )
        if err := writeAPI.WritePoint(ctx, point); err != nil {
            log.Printf("InfluxDB write failed: %v", err)
            continue
        }

        log.Printf("Wrote to InfluxDB: %s persons=%d cars=%d lat=%.1fms",
            det.Camera, det.Persons, det.Cars, det.LatencyMS,
        )
    }
}
