package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type ZoneStore struct {
	db *sql.DB
}

type Zone struct {
	ID              int64      `json:"id,omitempty"`
	Name            string     `json:"name,omitempty"`
	Camera          string     `json:"camera,omitempty"`
	Polygon         []Point    `json:"polygon,omitempty"`
	IsActive        bool       `json:"is_active"`
	ActiveFrom      *time.Time `json:"active_from,omitempty"`
	ActiveUntil     *time.Time `json:"active_until,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at,omitempty"`
	LoiterThreshold *int       `json:"loiter_threshold_seconds"`
	DefaultSeverity string     `json:"default_severity"`
}

type Point struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
}

func NewZonesStore(db *sql.DB) *ZoneStore {
	return &ZoneStore{db: db}
}

func (s *ZoneStore) Create(ctx context.Context, zone *Zone) (*Zone, error) {

	query := `
		INSERT INTO zones (name, camera, polygon, is_active, active_from, active_until, loiter_threshold_seconds, default_severity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, name, camera, polygon, is_active, active_from, active_until, updated_at, created_at, loiter_threshold_seconds, default_severity 
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonB, err := json.Marshal(zone.Polygon)
	if err != nil {
		return nil, err
	}

	var polygonRaw []byte
	err = s.db.QueryRowContext(ctx, query, zone.Name, zone.Camera, jsonB, zone.IsActive, zone.ActiveFrom, zone.ActiveUntil, zone.LoiterThreshold, zone.DefaultSeverity).
		Scan(&zone.ID, &zone.Name, &zone.Camera, &polygonRaw, &zone.IsActive, &zone.ActiveFrom, &zone.ActiveUntil, &zone.UpdatedAt, &zone.CreatedAt, &zone.LoiterThreshold, &zone.DefaultSeverity)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(polygonRaw, &zone.Polygon); err != nil {
		return nil, err
	}

	return zone, nil
}

func (s *ZoneStore) Get(ctx context.Context, camName string) ([]Zone, error) {

	query := `
		SELECT id, name, camera, polygon, is_active, active_from, active_until, updated_at, created_at, loiter_threshold_seconds, default_severity
		FROM zones
		WHERE camera = $1;
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, camName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	var jsonB []byte

	for rows.Next() {

		var row Zone

		if err := rows.Scan(&row.ID, &row.Name, &row.Camera, &jsonB, &row.IsActive, &row.ActiveFrom, &row.ActiveUntil, &row.UpdatedAt, &row.CreatedAt, &row.LoiterThreshold, &row.DefaultSeverity); err != nil {
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

func (s *ZoneStore) Update(ctx context.Context, id int64, zone *Zone) (*Zone, error) {

	query := `
		UPDATE zones
		SET name = $1,
		    camera = $2,
		    polygon = $3,
		    is_active = $4,
		    active_from = $5,
		    active_until = $6,
			loiter_threshold_seconds = $7,
			default_severity = $8,
			updated_at = NOW()
		WHERE id = $9
		RETURNING id, name, camera, polygon, is_active, active_from, active_until, updated_at, created_at, loiter_threshold_seconds, default_severity;
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonM, err := json.Marshal(zone.Polygon)
	if err != nil {
		return nil, err
	}

	var jsonB []byte
	err = s.db.QueryRowContext(ctx, query, zone.Name, zone.Camera, jsonM, zone.IsActive, zone.ActiveFrom, zone.ActiveUntil, zone.LoiterThreshold, zone.DefaultSeverity, id).
		Scan(&zone.ID,
			&zone.Name,
			&zone.Camera,
			&jsonB,
			&zone.IsActive,
			&zone.ActiveFrom,
			&zone.ActiveUntil,
			&zone.UpdatedAt,
			&zone.CreatedAt,
			&zone.LoiterThreshold,
			&zone.DefaultSeverity,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	if err := json.Unmarshal(jsonB, &zone.Polygon); err != nil {
		return nil, err
	}

	return zone, nil

}

func (s *ZoneStore) Delete(ctx context.Context, id int64) error {

	query := `
		DELETE FROM zones WHERE id = $1;
	`
	res, err := s.db.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	row, err := res.RowsAffected()

	if row == 0 {
		return sql.ErrNoRows
	}

	if err != nil {
		return err
	}

	return nil
}
