package main

import (
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
)

const uploadsTopic = "video.uploads"

func ensureUploadsTopic(broker string) {

	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		log.Printf("could not dial kafka to ensure %s topic: %v", uploadsTopic, err)
		return
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             uploadsTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Printf("could not create %s topic: %v", uploadsTopic, err)
	}
}
