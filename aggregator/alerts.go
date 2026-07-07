package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AlertStore struct {
	db *sql.DB
}

func NewAlertStore(db *sql.DB) *AlertStore {
	return &AlertStore{db: db}
}

func (s *AlertStore) Create(ctx context.Context, alerts []Alert) error {

	if len(alerts) == 0 {
		return nil
	}
	query := `
		INSERT INTO alerts (alert_id, camera, zone_id, tracker_id, bound_box, label, recorded_at, alert_type, severity)
		VALUES 
	`
	args := []interface{}{}
	values := []string{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for i, a := range alerts {
		base := i * 9

		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9,
		))

		jsonBoundBox, err := json.Marshal(a.BoundBox)
		if err != nil {
			return err
		}

		args = append(args, a.AlertID, a.Camera, a.ZoneID, a.TrackerID, jsonBoundBox, a.Label, a.RecordedAt, a.AlertType, a.Severity)
	}

	query += strings.Join(values, ",")

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}
