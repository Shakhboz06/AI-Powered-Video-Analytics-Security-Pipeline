CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS trackings(
    camera VARCHAR(255) NOT NULL,
    tracker_id INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    conf_score FLOAT NOT NULL,
    bound_box JSONB NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);

SELECT create_hypertable(
    'trackings', 
    by_range('recorded_at')
);



