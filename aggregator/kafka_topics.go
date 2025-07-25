
package main

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/segmentio/kafka-go"
)

type TopicSpec struct {
    Name              string
    NumPartitions     int
    ReplicationFactor int
}

func KafkaTopics(ctx context.Context, broker string, specs []TopicSpec) error {
    conn, err := kafka.DialContext(ctx, "tcp", broker)
    if err != nil {
        return fmt.Errorf("dial %s: %w", broker, err)
    }
    defer conn.Close()

    // 1) create topics
    var cfgs []kafka.TopicConfig
    for _, s := range specs {
        cfgs = append(cfgs, kafka.TopicConfig{
            Topic:             s.Name,
            NumPartitions:     s.NumPartitions,
            ReplicationFactor: s.ReplicationFactor,
        })
    }
    if err := conn.CreateTopics(cfgs...); err != nil {
        if !strings.Contains(err.Error(), "already exists") {
            return fmt.Errorf("create topics: %w", err)
        }
    }

    // 2) wait for leaders
    for _, s := range specs {
        deadline := time.Now().Add(30 * time.Second)
        for time.Now().Before(deadline) {
            parts, _ := conn.ReadPartitions(s.Name)
            ready := true
            for _, p := range parts {
                if p.Leader.ID == -1 {
                    ready = false
                    break
                }
            }
            if ready {
                break
            }
            time.Sleep(500 * time.Millisecond)
        }
    }
    return nil
}
