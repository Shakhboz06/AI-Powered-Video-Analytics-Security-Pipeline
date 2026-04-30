package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type DetectionsList struct {
	Camera          []string         `json:"camera,omitempty"`
	TotalDetections []int64          `json:"total_detections,omitempty"`
	TotalObjects    []map[string]int `json:"total_objects,omitempty"`
	LatencyMS       []float64        `json:"latency_ms,omitempty"`
	RecordedAt      []time.Time      `json:"recorded_at,omitempty"`
}

type Storage struct{
	Users interface{
		Get(context.Context, int64)(*Users, error)
		Set(context.Context,  *Users) error
	}
	Detections interface{
		Get(context.Context, int64)(*DetectionsList, error)
		Set(context.Context,  *DetectionsList) error
	}
}
func NewRedisStorage(rbd *redis.Client)Storage {
	return Storage{
		Users: &UsersStore{rbd: rbd},
		// Detections: DetectionsStore{rbd: rbd},
	}
}