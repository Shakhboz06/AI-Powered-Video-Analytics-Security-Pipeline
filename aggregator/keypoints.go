package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type KeypointStore struct {
	db *sql.DB
}

func NewKeypointStore(db *sql.DB) *KeypointStore {
	return &KeypointStore{db: db}
}

type Keypoints struct {
	TrackerID int64  `json:"tracker_id"`
	Points    Points `json:"points"`
}

type Points struct {
	XY   [][]float64 `json:"xy"`
	Conf []float64   `json:"conf"`
}

func (s *KeypointStore) CreateKeyPoints(ctx context.Context, camera string, recorded_at int64, kyp []Keypoints) error {

	if len(kyp) == 0 {
		return nil
	}
	query := `
		INSERT INTO keypoints (camera, tracker_id, recorded_at, keypoint)
		VALUES 
	`
	args := []interface{}{}
	values := []string{}

	for i, key := range kyp {
		base := i * 4

		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4,
		))

		jsonKeyPoints, err := json.Marshal(key.Points)
		if err != nil {
			return err
		}

		args = append(args, camera, key.TrackerID, time.UnixMilli(recorded_at), jsonKeyPoints)
	}

	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	query += strings.Join(values, ",")
	
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}
