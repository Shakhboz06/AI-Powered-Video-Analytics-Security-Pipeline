package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type ZoneStore struct {
	db *sql.DB
}

func NewZoneStore(db *sql.DB) *ZoneStore {
	return &ZoneStore{db: db}
}

func (s *ZoneStore) FetchAllZones(ctx context.Context) ([]Zone, error) {
	query := `
		SELECT id, name, camera, polygon, is_active, active_from, active_until, loiter_threshold_seconds, default_severity
		FROM zones 
		WHERE 
		is_active = true;
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	var jsonB []byte

	for rows.Next() {

		var row Zone

		if err := rows.Scan(&row.ID, &row.Name, &row.Camera, &jsonB, &row.IsActive, &row.ActiveFrom, &row.ActiveUntil, &row.LoiterThreshold, &row.DefaultSeverity); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonB, &row.Polygon); err != nil {
			return nil, err
		}

		zones = append(zones, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return zones, nil

}
