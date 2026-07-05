ALTER TABLE alerts ADD COLUMN alert_id UUID;

CREATE UNIQUE INDEX IF NOT EXISTS alerts_alert_id_recorded_at_unique_idx
ON alerts (alert_id, recorded_at);