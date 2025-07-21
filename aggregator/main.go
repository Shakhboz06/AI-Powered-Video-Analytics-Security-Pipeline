package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
	"video-analytics-pipe/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	// "github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

type Result struct {
	Camera    string  `json:"camera"`
	Persons   int  `json:persons`
	Cars      int  `json:cars`
	LatencyMS float64 `json:"latency_ms"`
}

func retryKafka(broker string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", broker, 1*time.Second)
		if err == nil {
			conn.Close()
			log.Printf("kafka is up at %s", broker)
			return
		}
		log.Printf("Waiting for Kafka at %s… (%v)", broker, err)
		time.Sleep(1 * time.Second)
	}
	log.Fatalf("timeout waiting for Kafka at %s", broker)
}
func main() {

	// if err := godotenv.Load("../.env"); err != nil {
	//     log.Printf("could not load .env: %v", err)
	// }

	influxURL := config.GetString("INFLUX_URL", "http://localhost:8086")

	token := config.GetString("INFLUX_TOKEN", "")
	org := config.GetString("INFLUX_ORG", "")
	bucket := config.GetString("INFLUX_BUCKET", "")

	fmt.Printf("Connecting to %s; org=%s, bucket=%s\n", influxURL, org, bucket)

	client := influxdb2.NewClient(influxURL, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)
	defer client.Close()

	broker := config.GetString("KAFKA_BROKER", "kafka:9092")
	topic := config.GetString("KAFKA_TOPIC", "video.results")

	retryKafka(broker, 60*time.Second)

	KafkaTopics(broker, []string{"video.analysis", "video.results"})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "aggregator-group",
	})
	defer reader.Close()

	for {
		mes, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			time.Sleep(time.Second)
			continue
		}

		var res Result
		if err := json.Unmarshal(mes.Value, &res); err != nil {
			fmt.Printf("Invalid JSON (%s): %v\n", string(mes.Value), err)
			continue
		}

		p := influxdb2.NewPoint(
			"frame_detections",
			map[string]string{"camera": res.Camera},
			map[string]interface{}{
				"persons": res.Persons, 
				"cars": res.Cars,
				"latency_ms": res.LatencyMS,
			},
			time.Now(),
		)
		if err := writeAPI.WritePoint(context.Background(), p); err != nil {
			fmt.Printf("Failed write: %v\n", err)
			continue
		}
		fmt.Printf( "Wrote %s (persons=%d, cars=%d latency=%.1fms)\n", res.Camera, res.Persons, res.Cars, res.LatencyMS,)
	}

}
