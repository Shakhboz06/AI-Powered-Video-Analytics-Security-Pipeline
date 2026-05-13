CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS keypoints (
    camera TEXT NOT NULL,
    tracker_id INT NOT NULL,
    keypoint JSONB NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);

SELECT create_hypertable('keypoints', 'recorded_at');