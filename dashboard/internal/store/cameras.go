package store

import (
	"context"
	"database/sql"
	"time"
)

type Cameras struct {
	ID          int64     `json:"camera_id"`
	Name        string    `json:"camera_name"`
	VideoSource string    `json:"video_source"`
	IsActive    bool      `json:"is_active"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type CameraStore struct {
	db *sql.DB
}

func NewCameraStore(db *sql.DB) *CameraStore{
	return &CameraStore{db: db}
}

func (s *CameraStore) Create(ctx context.Context, cameras *Cameras) (*Cameras, error) {

	query := `INSERT INTO cameras (camera_name, video_source, is_active)
		VALUES ($1, $2, $3) RETURNING camera_id, camera_name, video_source, is_active, updated_at, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var camera Cameras
	err := s.db.QueryRowContext(ctx, query, cameras.Name, cameras.VideoSource, cameras.IsActive).
		Scan(&camera.ID, &camera.Name, &camera.VideoSource, &camera.IsActive, &camera.UpdatedAt, &camera.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &camera, nil
}

func (s *CameraStore) GetAll(ctx context.Context) ([]Cameras, error) {

	query := `SELECT camera_id, camera_name, video_source, is_active, updated_at, created_at
		FROM cameras
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var cameras []Cameras
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var cam Cameras

		err := rows.Scan(&cam.ID, &cam.Name, &cam.VideoSource, &cam.IsActive, &cam.UpdatedAt, &cam.CreatedAt)

		if err != nil {
			return nil, err
		}

		cameras = append(cameras, cam)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cameras, nil
}


func (s *CameraStore) Update (ctx context.Context, camera Cameras, id int64)(*Cameras, error){
	
	query := `UPDATE cameras
		SET camera_name = $1, 
		video_source = $2, 
		is_active = $3,
		updated_at = NOW()
		WHERE camera_id = $4
		RETURNING camera_id, camera_name, video_source, is_active, updated_at, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	
	err := s.db.QueryRowContext(ctx, query, camera.Name, camera.VideoSource, camera.IsActive, id).
	Scan(&camera.ID, &camera.Name, &camera.VideoSource, &camera.IsActive, &camera.UpdatedAt, &camera.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &camera, nil
}

func (s *CameraStore) Delete(ctx context.Context, id int64) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	query := `
		DELETE FROM cameras WHERE camera_id = $1;
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
