CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS detections(
    camera VARCHAR(255) NOT NULL,
    latency_ms DOUBLE PRECISION NOT NULL,
    total_objects JSONB NOT NULL,
    total_detections INTEGER NOT NULL,
    detections JSONB NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);

SELECT create_hypertable(
    'detections', 
    by_range('recorded_at')
);



