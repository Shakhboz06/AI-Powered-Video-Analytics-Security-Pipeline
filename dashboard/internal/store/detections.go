package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type DetectionStore struct {
	db *sql.DB
}

type DetectionsList struct {
	Camera          []string         `json:"camera,omitempty"`
	TotalDetections []int64          `json:"total_detections,omitempty"`
	TotalObjects    []map[string]int `json:"total_objects,omitempty"`
	LatencyMS       []float64        `json:"latency_ms,omitempty"`
	RecordedAt      []time.Time      `json:"recorded_at,omitempty"`
}

type Detections struct {
	Camera          string         `json:"camera,omitempty"`
	TotalDetections int64          `json:"total_detections,omitempty"`
	TotalObjects    map[string]int `json:"total_objects,omitempty"`
	LatencyMS       float64        `json:"latency_ms,omitempty"`
	RecordedAt      time.Time      `json:"recorded_at,omitempty"`
}

type AggSummary struct {
	Bucket             []time.Time `json:"bucket,omitempty"`
	AvgTotalDetections []float64   `json:"avg_total_detections,omitempty"`
	MaxTotalDetections []int64     `json:"max_total_detections,omitempty"`
	SumTotalDetections []int64     `json:"sum_total_detections,omitempty"`
	AvgLatencyMS       []float64   `json:"avg_latency_ms,omitempty"`
}

func NewDetStore(db *sql.DB) *DetectionStore {
	return &DetectionStore{db: db}
}

func (s *DetectionStore) GetAllCamera(ctx context.Context) (*DetectionsList, error) {

	// query := `
	// 	SELECT DISTINCT camera
	// 	FROM detections
	// 	WHERE recorded_at BETWEEN NOW() - INTERVAL '1 hour' AND NOW();
	// `

	query := `
		SELECT DISTINCT camera
		FROM detections
		WHERE recorded_at BETWEEN NOW() - INTERVAL '24 hours' AND NOW();
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var detection DetectionsList
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var camera string

		if err := rows.Scan(&camera); err != nil {
			return nil, err
		}
		detection.Camera = append(detection.Camera, camera)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &detection, err
}

func (s *DetectionStore) GetDetectionCounts(ctx context.Context, cameraName string, timeRange time.Time) (*DetectionsList, error) {

	// SELECT recorded_at, counts, latency_ms, total_detections
	// FROM detections
	// WHERE camera = $1
	//   AND recorded_at BETWEEN $2 AND $3
	// ORDER BY recorded_at ASC

	query := `
			SELECT recorded_at, total_objects, latency_ms, total_detections
    		FROM detections
    		WHERE camera = $1
      		AND recorded_at >= $2
			ORDER BY recorded_at ASC
			`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, cameraName, timeRange)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var detections DetectionsList

	for rows.Next() {

		var totalDetection int64
		var totalObjects map[string]int
		var latencyMS float64
		var recordedAt time.Time
		var jsobnB []byte

		if err := rows.Scan(&recordedAt, &jsobnB, &latencyMS, &totalDetection); err != nil {
			return nil, err
		}

		err := json.Unmarshal(jsobnB, &totalObjects)
		if err != nil {
			return nil, err
		}

		detections.TotalDetections = append(detections.TotalDetections, totalDetection)
		detections.LatencyMS = append(detections.LatencyMS, latencyMS)
		detections.TotalObjects = append(detections.TotalObjects, totalObjects)
		detections.RecordedAt = append(detections.RecordedAt, recordedAt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &detections, nil
}

func (s *DetectionStore) GetLiveByCam(ctx context.Context, cameraName string) (*Detections, error) {

	query := `
		SELECT recorded_at, total_objects, latency_ms, total_detections
		FROM detections
		WHERE camera = $1
		AND recorded_at >= NOW() - INTERVAL '24 hours'
  		ORDER BY recorded_at DESC LIMIT 1;
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var detection Detections

	var jsobnB []byte

	err := s.db.QueryRowContext(ctx, query, cameraName).Scan(&detection.RecordedAt, &jsobnB, &detection.LatencyMS, &detection.TotalDetections)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsobnB, &detection.TotalObjects); err != nil {
		return nil, err
	}

	return &detection, nil

}

func (s *DetectionStore) GetAggSummary(ctx context.Context, cameraName, bucket string, startTime, endTime time.Time) (*AggSummary, error) {

	query := `
			SELECT
  			time_bucket($1 * INTERVAL '1 second', recorded_at) AS bucket,
  			AVG(total_detections)::double precision AS avg_total_detections,
  			MAX(total_detections)                   AS max_total_detections,
  			SUM(total_detections)                   AS sum_total_detections,
  			AVG(latency_ms)::double precision       AS avg_latency_ms
			FROM detections
			WHERE camera = $2
  			AND recorded_at >= $3::timestamptz
  			AND recorded_at <  $4::timestamptz
			GROUP BY bucket
			ORDER BY bucket ASC;
			`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	bucketDuration, err := time.ParseDuration(bucket)
	if err != nil {
		log.Println("cannot parse the bucket duration")
		return nil, err
	}

	bucketSeconds := int64(bucketDuration / time.Second)

	rows, err := s.db.QueryContext(ctx, query, bucketSeconds, cameraName, startTime, endTime)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var summary AggSummary

	for rows.Next() {

		var bucketSize time.Time
		var avgTotalDetections float64
		var maxTotalDetections int64
		var sumTotalDetections int64
		var avgLatencyMS float64

		if err := rows.Scan(&bucketSize, &avgTotalDetections, &maxTotalDetections, &sumTotalDetections, &avgLatencyMS); err != nil {
			return nil, err
		}

		summary.Bucket = append(summary.Bucket, bucketSize)
		summary.AvgTotalDetections = append(summary.AvgTotalDetections, avgTotalDetections)
		summary.MaxTotalDetections = append(summary.MaxTotalDetections, maxTotalDetections)
		summary.SumTotalDetections = append(summary.SumTotalDetections, sumTotalDetections)
		summary.AvgLatencyMS = append(summary.AvgLatencyMS, avgLatencyMS)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}


	return &summary, nil
}

func (s *DetectionStore) GetClassSum(ctx context.Context, cameraName string, startTime, endTime time.Time) (map[string]int64, error) {

	query := `
			SELECT
			kv.key AS class_name,
			SUM(kv.value::bigint) AS total_count
			FROM detections d
			CROSS JOIN LATERAL jsonb_each_text(d.total_objects) AS kv(key, value)
			WHERE d.camera = $1
			AND d.recorded_at >= $2::timestamptz
			AND d.recorded_at <  $3::timestamptz
			GROUP BY kv.key
			ORDER BY total_count DESC, kv.key ASC;
			`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()


	rows, err := s.db.QueryContext(ctx, query, cameraName, startTime, endTime)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	classSum := map[string]int64{}

	for rows.Next() {
		
		var className string
		var totalCount int64
		
		
		if err := rows.Scan(&className, &totalCount); err != nil {
			return nil, err
		}
		
		classSum[className] = totalCount	
	}


	if err := rows.Err(); err != nil {
		return nil, err
	}


	return classSum, nil
}

func (s *DetectionStore) GetByLatencyMS(ctx context.Context, cameraName string, startTime, endTime time.Time, threshold float64) (*DetectionsList, error) {

	query := `
			SELECT
			  recorded_at,
			  camera,
			  latency_ms,
			  total_detections
			FROM detections
			WHERE camera = $1
			  AND recorded_at >= $2::timestamptz
			  AND recorded_at <  $3::timestamptz
			  AND latency_ms > $4
			ORDER BY latency_ms DESC, recorded_at DESC;
			`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, cameraName, startTime, endTime,threshold)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var detections DetectionsList

	for rows.Next() {

		var camera string
		var totalDetection int64
		var latencyMS float64
		var recordedAt time.Time

		if err := rows.Scan(&recordedAt, &camera, &latencyMS, &totalDetection); err != nil {
			return nil, err
		}

		detections.Camera = append(detections.Camera, camera)
		detections.RecordedAt = append(detections.RecordedAt, recordedAt)
		detections.LatencyMS = append(detections.LatencyMS, latencyMS)
		detections.TotalDetections = append(detections.TotalDetections, totalDetection)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &detections, nil
}
