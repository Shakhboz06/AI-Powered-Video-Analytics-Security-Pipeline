package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type ZoneCache struct {
	mu    sync.RWMutex
	zones map[string][]Zone
}

type Zone struct {
	ID              int64      `json:"id,omitempty"`
	Name            string     `json:"name,omitempty"`
	Camera          string     `json:"camera,omitempty"`
	Polygon         []Point    `json:"polygon,omitempty"`
	IsActive        bool       `json:"is_active"`
	ActiveFrom      *time.Time `json:"active_from,omitempty"`
	ActiveUntil     *time.Time `json:"active_until,omitempty"`
	LoiterThreshold *int       `json:"loiter_threshold_seconds"`
	DefaultSeverity string     `json:"default_severity"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func NewZoneCache() *ZoneCache {
	return &ZoneCache{
		zones: make(map[string][]Zone),
	}
}

func (c *ZoneCache) Get(camera string) []Zone {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.zones[camera]
}

func (c *ZoneCache) Refresh(allZones []Zone) {

	newMap := make(map[string][]Zone)

	for _, z := range allZones {
		newMap[z.Camera] = append(newMap[z.Camera], z)
	}

	c.mu.Lock()
	c.zones = newMap
	c.mu.Unlock()
}

func (c *ZoneCache) BackgroundRefresh(ctx context.Context, interval time.Duration, store *ZoneStore) error {

	zones, err := store.FetchAllZones(ctx)
	if err != nil {
		return err
	}

	c.Refresh(zones)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				zones, err := store.FetchAllZones(ctx)
				if err != nil {
					log.Printf("zone cache refresh failed: %v", err)
					continue
				}
				c.Refresh(zones)
			}
		}
	}()

	return nil
}
