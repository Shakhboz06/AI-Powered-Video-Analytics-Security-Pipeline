package store

import (
	"context"
	"database/sql"
	"time"
)

type JobStore struct {
	db *sql.DB
}

type AnalysisJob struct {
	ID               int64      `json:"-"`
	JobID            string     `json:"job_id"`
	StreamID         string     `json:"stream_id"`
	OriginalFilename string     `json:"original_filename"`
	Status           string     `json:"status"`
	Progress         int        `json:"progress"`
	Error            *string    `json:"error,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	LastFrameAt      *time.Time `json:"last_frame_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func NewJobStore(db *sql.DB) *JobStore {
	return &JobStore{db: db}
}

func (s *JobStore) Create(ctx context.Context, job *AnalysisJob) error {

	query := `
		INSERT INTO analysis_jobs (job_id, stream_id, original_filename, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.db.QueryRowContext(ctx, query, job.JobID, job.StreamID, job.OriginalFilename, job.Status).
		Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)
}

func (s *JobStore) GetByJobID(ctx context.Context, jobID string) (*AnalysisJob, error) {

	query := `
		SELECT id, job_id, stream_id, original_filename, status, progress, error,
		       started_at, last_frame_at, created_at, updated_at
		FROM analysis_jobs
		WHERE job_id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var job AnalysisJob
	err := s.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.JobID, &job.StreamID, &job.OriginalFilename, &job.Status,
		&job.Progress, &job.Error, &job.StartedAt, &job.LastFrameAt,
		&job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// UpdateProgress is called by the upload ingestor while it streams frames
// into the pipeline. started_at is stamped on the first "processing" update
// and last_frame_at records the timestamp of the final published frame so
// completion can be detected against the detections written by the worker.
func (s *JobStore) UpdateProgress(ctx context.Context, jobID, status string, progress int, errMsg *string, lastFrameAt *time.Time) error {

	query := `
		UPDATE analysis_jobs
		SET status = $2,
		    progress = GREATEST(progress, $3),
		    error = COALESCE($4, error),
		    started_at = CASE WHEN started_at IS NULL AND $2 <> 'queued' THEN NOW() ELSE started_at END,
		    last_frame_at = COALESCE($5, last_frame_at),
		    updated_at = NOW()
		WHERE job_id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, jobID, status, progress, errMsg, lastFrameAt)
	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// FinalizeIfComplete flips a "finalizing" job to "done" once the worker has
// caught up: the newest detection row for the job's stream has reached the
// timestamp of the last ingested frame (with a small tolerance). A stale
// fallback marks the job done anyway so a link never spins forever.
func (s *JobStore) FinalizeIfComplete(ctx context.Context, job *AnalysisJob) error {

	if job.Status != "finalizing" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	done := false

	if job.LastFrameAt != nil {
		var latest sql.NullTime
		err := s.db.QueryRowContext(ctx,
			`SELECT MAX(recorded_at) FROM detections WHERE camera = $1`,
			job.StreamID,
		).Scan(&latest)
		if err != nil {
			return err
		}

		if latest.Valid && !latest.Time.Before(job.LastFrameAt.Add(-2*time.Second)) {
			done = true
		}
	}

	// Fallback: worker lag should never hold a shared link hostage.
	if !done && time.Since(job.UpdatedAt) > 10*time.Minute {
		done = true
	}

	if !done {
		return nil
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE analysis_jobs SET status = 'done', progress = 100, updated_at = NOW() WHERE job_id = $1 AND status = 'finalizing'`,
		job.JobID,
	)
	if err != nil {
		return err
	}

	job.Status = "done"
	job.Progress = 100
	return nil
}
