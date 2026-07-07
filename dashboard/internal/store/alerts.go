package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type AlertStore struct {
	db *sql.DB
}

type Alerts struct {
	ID         int64      `json:"id"`
	AlertId    string     `json:"alert_id"`
	Camera     string     `json:"camera"`
	ZoneName   *string    `json:"zone_name"`
	ZoneID     *int64     `json:"zone_id"`
	TrackerID  int64      `json:"tracker_id"`
	BoundBox   [4]float64 `json:"bound_box"`
	Label      string     `json:"label"`
	Status     string     `json:"status"`
	RecordedAt time.Time  `json:"recorded_at"`
	AlertType  string     `json:"alert_type"`
	Severity   string     `json:"severity"`
}

type AlertFrames struct{
	FrameUUID string `json:"alert_id"`
}

func NewAlertStore(db *sql.DB) *AlertStore {
	return &AlertStore{db: db}
}

func (s *AlertStore) GetAll(ctx context.Context, camera, status string) ([]Alerts, error) {

	query := `SELECT 
    a.id,
	a.alert_id,
    a.camera,
    a.zone_id,
    z.name AS zone_name,
    a.tracker_id,
    a.bound_box,
    a.label,
    a.status,
    a.recorded_at,
	a.alert_type,
	a.severity
	FROM alerts a
	LEFT JOIN zones z ON a.zone_id = z.id
	WHERE (a.camera = $1 OR $1 IS NULL)
	AND (a.status = $2 OR $2 IS NULL)
	AND a.recorded_at >= NOW() - INTERVAL '24 hours'
	ORDER BY a.recorded_at DESC
	LIMIT 100;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var cameraArg sql.NullString
	if camera != "" {
		cameraArg = sql.NullString{String: camera, Valid: true}
	}

	var statusArg sql.NullString
	if status != "" {
		statusArg = sql.NullString{String: status, Valid: true}
	}

	rows, err := s.db.QueryContext(ctx, query, cameraArg, statusArg)
	if err != nil {
		return nil, err
	}

	var alerts []Alerts

	for rows.Next() {

		var a Alerts

		var rowBoundBox []byte
		if err := rows.Scan(&a.ID, &a.AlertId, &a.Camera, &a.ZoneID, &a.ZoneName, &a.TrackerID, &rowBoundBox, &a.Label, &a.Status, &a.RecordedAt, &a.AlertType, &a.Severity); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rowBoundBox, &a.BoundBox); err != nil {
			return nil, err
		}

		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (s *AlertStore) UpdateStatus(ctx context.Context, id int64, status string) (*Alerts, error) {

	query := `
		WITH updated AS (
    	UPDATE alerts
    	SET status = $2
    	WHERE id = $1
    	RETURNING id, alert_id, camera, zone_id, tracker_id, bound_box, label, status, recorded_at, alert_type, severity
		)
		SELECT u.id, u.camera, u.zone_id, z.name AS zone_name,
    		u.tracker_id, u.bound_box, u.label, u.status, u.recorded_at, u.alert_type, u.severity
		FROM updated u
		LEFT JOIN zones z ON u.zone_id = z.id;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var alert Alerts
	var rowBoundBox []byte

	err := s.db.QueryRowContext(ctx, query, id, status).Scan(&alert.ID, &alert.AlertId, &alert.Camera, &alert.ZoneID, &alert.ZoneName, &alert.TrackerID, &rowBoundBox, &alert.Label, &alert.Status, &alert.RecordedAt, &alert.AlertType, &alert.Severity)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rowBoundBox, &alert.BoundBox); err != nil {
		return nil, err
	}

	return &alert, nil
}

func (s *AlertStore) GetForStream(ctx context.Context, camera string) ([]Alerts, error) {

	query := `SELECT
    a.id,
	a.alert_id,
    a.camera,
    a.zone_id,
    z.name AS zone_name,
    a.tracker_id,
    a.bound_box,
    a.label,
    a.status,
    a.recorded_at,
	a.alert_type,
	a.severity
	FROM alerts a
	LEFT JOIN zones z ON a.zone_id = z.id
	WHERE a.camera = $1
	ORDER BY a.recorded_at ASC
	LIMIT 500;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, camera)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []Alerts{}

	for rows.Next() {

		var a Alerts

		var rowBoundBox []byte
		if err := rows.Scan(&a.ID, &a.AlertId, &a.Camera, &a.ZoneID, &a.ZoneName, &a.TrackerID, &rowBoundBox, &a.Label, &a.Status, &a.RecordedAt, &a.AlertType, &a.Severity); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rowBoundBox, &a.BoundBox); err != nil {
			return nil, err
		}

		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}


func (s *AlertStore) AlertBelongsToStream(ctx context.Context, alertID, camera string) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM alerts WHERE alert_id = $1 AND camera = $2)`,
		alertID, camera,
	).Scan(&exists)

	return exists, err
}
