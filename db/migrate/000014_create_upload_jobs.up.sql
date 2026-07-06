CREATE TABLE IF NOT EXISTS upload_jobs(
    id BIGSERIAL PRIMARY KEY,
    job_id UUID NOT NULL UNIQUE,
    filename VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- job_id is indexed by its UNIQUE constraint; no extra index needed.
