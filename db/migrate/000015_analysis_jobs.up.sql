CREATE TABLE IF NOT EXISTS analysis_jobs(
    id BIGSERIAL PRIMARY KEY,
    job_id UUID NOT NULL UNIQUE,
    stream_id VARCHAR(255) NOT NULL UNIQUE,
    original_filename VARCHAR(512) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    progress SMALLINT NOT NULL DEFAULT 0,
    error TEXT,
    started_at TIMESTAMPTZ,
    last_frame_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS analysis_jobs_job_id_idx ON analysis_jobs (job_id);
