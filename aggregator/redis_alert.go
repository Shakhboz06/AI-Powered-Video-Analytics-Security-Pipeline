package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

func publishAlertsToRedis(ctx context.Context, rdb *redis.Client, alerts []Alert) {
	for _, a := range alerts {
		payload, err := json.Marshal(a)
		if err != nil {
			log.Printf("marshal alert: %v", err)
			continue
		}

		if err := rdb.Publish(ctx, "alerts", payload).Err(); err != nil {
			log.Printf("published alert: %v", err)
		}

	}
}
