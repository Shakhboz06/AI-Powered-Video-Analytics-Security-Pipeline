CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS alerts(
    id BIGSERIAL NOT NULL,
    camera VARCHAR(255) NOT NULL,
    zone_id BIGINT REFERENCES zones(id),
    tracker_id BIGINT NOT NULL,
    bound_box JSONB NOT NULL,
    label VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    recorded_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, recorded_at)
);

SELECT create_hypertable(
    'alerts', 
    by_range('recorded_at')
);