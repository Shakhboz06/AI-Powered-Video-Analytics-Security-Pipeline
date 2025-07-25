package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"video-analytics-pipe/config"

	"github.com/segmentio/kafka-go"
)

func main() {

	broker := config.GetString("KAFKA_BROKER", "kafka:9092")
	inTopic := config.GetString("KAFKA_RESULTS_TOPIC", "video.results")
	outTopic := config.GetString("KAFKA_ENRICHED_TOPIC", "video.enriched")
	groupID := config.GetString("BROKER_GROUP_ID", "broker-group")

	// reader (consumer)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic: inTopic,
		GroupID: groupID,
	})
	defer reader.Close()

	// writer(producer)
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic: outTopic,
		Balancer: &kafka.LeastBytes{},
	})

	defer writer.Close()

	log.Printf("Broker: consuming %q → producing %q\n", inTopic, outTopic)

	for{
		message, err := reader.ReadMessage(context.Background())
		if err != nil{
			fmt.Printf("read error: %v\n", err)
			time.Sleep(time.Second)
			continue
		}

		// unmarshal the raw result
		var record map[string]any
		if err := json.Unmarshal(message.Value, &record); err != nil{
			fmt.Printf("invalid JSON %q: %v\n", message.Value, err)
            continue
		}


		// ----- ENRICHMENT STUB -----
        // e.g. convert timestamp to RFC3339, add `region` tag, filter low confidence

		if record["latency_ms"].(float64) > 500{
			// skip high-latency records, or tag them
			record["flag"] = "high_latency"
		}

		// convert ms timestamp to ISO time
        if ts, ok := record["timestamp"].(float64); ok {
            record["timestamp_iso"] = time.UnixMilli(int64(ts)).Format(time.RFC3339)
        }


		out, err := json.Marshal(record)
		if err != nil{
			fmt.Printf("marshalling error:", err)
			continue
		}

		// write enriched record downstream
		if err := writer.WriteMessages((context.Background()), kafka.Message{Value: out}); err != nil{
			fmt.Printf("write error: %v\n", err)
		}
		
	}


}