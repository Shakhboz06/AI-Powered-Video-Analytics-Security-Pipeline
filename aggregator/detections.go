package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewDetStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateDetections(ctx context.Context, detection *TabDetection) error {

	query := `
		INSERT INTO detections (camera, latency_ms, total_objects, total_detections, detections, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, detection.Camera, detection.LatencyMS, detection.TotalObjects, detection.TotalDetections, detection.Parameters, detection.RecordedAt)

	return err
}

func (s *Store) CreateTrackings(ctx context.Context, camera string, recorded_at int64, trackings *[]Detections) error {
	
	if trackings == nil || len(*trackings) == 0 {
		return nil
	}

	query := `
		INSERT INTO trackings (camera, label, conf_score, bound_box, tracker_id, recorded_at)
		VALUES
	`

	args := []interface{}{}
	values := []string{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for i, t := range *trackings {

		base := i * 6

		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6,
		))

		jsonBoundBox, err := json.Marshal(t.BoundBox)
		if err != nil {
			return err
		}

		args = append(args, camera, t.Label, t.ConfScore, jsonBoundBox, t.TrackerID, time.UnixMilli(recorded_at))
	}

	query += strings.Join(values, ",")

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return err
}
