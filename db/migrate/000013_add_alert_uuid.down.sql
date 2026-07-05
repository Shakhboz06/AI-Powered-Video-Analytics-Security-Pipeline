DROP INDEX IF EXISTS alerts_alert_id_recorded_at_unique_idx;

ALTER TABLE alerts DROP COLUMN IF EXISTS alert_id;