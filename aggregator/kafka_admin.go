package main

import (
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func KafkaTopics(broker string, topics []string) {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		log.Fatalf("cannot connect to Kafka at %s: %v", broker, err)
	}
	defer conn.Close()

	// 1) Create topics
	var configs []kafka.TopicConfig
	for _, t := range topics {
		configs = append(configs, kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}
	if err := conn.CreateTopics(configs...); err != nil {
		// ignore “already exists”
		if !strings.Contains(err.Error(), "Topic with this name already exists") {
			log.Fatalf("⚠️  failed creating topics: %v", err)
		}
	}

	
	for _, t := range topics {
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			parts, err := conn.ReadPartitions(t)
			if err != nil {
				log.Fatalf("unable to read partitions for %s: %v", t, err)
			}
			ready := true
			for _, p := range parts {
				if p.Leader.ID == -1 {
					ready = false
					break
				}
			}
			if ready {
				log.Printf("topic %s is ready", t)
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
