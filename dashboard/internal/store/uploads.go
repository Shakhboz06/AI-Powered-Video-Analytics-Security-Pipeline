package store

import (
	"context"
	"database/sql"
	"time"
)

type UploadStore struct {
	db *sql.DB
}

type UploadJob struct {
	ID          int64      `json:"-"`
	JobID       string     `json:"job_id"`
	Filename    string     `json:"filename"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func NewUploadStore(db *sql.DB) *UploadStore {
	return &UploadStore{db: db}
}

func (s *UploadStore) Create(ctx context.Context, jobID, filename string) error {

	query := `
		INSERT INTO uploaded_jobs (job_id, filename, status)
		VALUES ($1, $2, 'queued');
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, jobID, filename)
	return err
}

func (s *UploadStore) UpdateStatus(ctx context.Context, jobID, status string) error {

	query := `
		UPDATE uploaded_jobs
		SET status = $2,
		    completed_at = CASE WHEN $2 IN ('done', 'failed') THEN NOW() ELSE completed_at END
		WHERE job_id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, jobID, status)
	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *UploadStore) Get(ctx context.Context, jobID string) (*UploadJob, error) {

	query := `
		SELECT id, job_id, filename, status, created_at, completed_at
		FROM uploaded_jobs
		WHERE job_id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var job UploadJob
	err := s.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.JobID, &job.Filename, &job.Status, &job.CreatedAt, &job.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}
