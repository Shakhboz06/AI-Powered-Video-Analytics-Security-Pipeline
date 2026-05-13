package main

import (
	"log"
	"math"
	"time"
)

type PerPersonFallState struct {
	BboxHistory       []BboxSnapshot
	HeadYHistory      []HeadYSnapshot
	ConfirmationCount int64
	IsFallen          bool
	lastAlerted       time.Time
	LastSeen          time.Time
}

type HeadYSnapshot struct {
	Timestamp time.Time
	Y         float64
}

type BboxSnapshot struct {
	Timestamp time.Time
	Width     float64
	Height    float64
}
type FallDetector struct {
	States      map[string]map[int64]*PerPersonFallState
	LastCleanup time.Time
}

const (
	ratioThresholdStanding float64       = 0.5
	ratioThresholdFallen   float64       = 0.8
	keypointConfThreshold  float64       = 0.3
	confirmationFrames     int64         = 5
	yVarianceThreshold     float64       = 50
	cooldownSeconds        time.Duration = 30 * time.Second
	historyLookbackFrames  int64         = 10
	normalizedThreshold    float64       = 0.15
	velocityThreshold      float64       = 100
)

func NewFallDetector() *FallDetector {
	return &FallDetector{
		States:      make(map[string]map[int64]*PerPersonFallState),
		LastCleanup: time.Now(),
	}
}

func (d *FallDetector) DetectFall(camera string, detections []Detections, keypoints []Keypoints, timestamp time.Time) []Alert {

	if timestamp.Sub(d.LastCleanup) >= 10*time.Second {
		d.cleanupOldStates(timestamp, 60*time.Second)
		d.LastCleanup = timestamp
	}
	var alerts []Alert
	keypointMap := make(map[int64]Keypoints)

	for _, kp := range keypoints {
		keypointMap[kp.TrackerID] = kp
	}

	for _, det := range detections {

		if det.Label != "person" {
			continue
		}

		keypoints, exist := keypointMap[det.TrackerID]
		if !exist {
			continue
		}

		if d.States[camera] == nil {
			d.States[camera] = make(map[int64]*PerPersonFallState)
		}

		state, exists := d.States[camera][det.TrackerID]
		if !exists {
			state = &PerPersonFallState{}
			d.States[camera][det.TrackerID] = state
		}
		state.LastSeen = timestamp

		alert := checkFalling(camera, state, det.BoundBox, keypoints, timestamp)

		if alert != nil {
			alerts = append(alerts, *alert)
		}

	}

	return alerts
}

func checkFalling(camera string, st *PerPersonFallState, bbox [4]float64, kyp Keypoints, timestamp time.Time) *Alert {

	width := bbox[2] - bbox[0]
	height := bbox[3] - bbox[1]

	if height <= 0 {
		return nil
	}

	currentRatio := width / height

	st.BboxHistory = append(st.BboxHistory, BboxSnapshot{
		Timestamp: timestamp,
		Width:     width,
		Height:    height,
	})

	if len(st.BboxHistory) > int(historyLookbackFrames) {
		st.BboxHistory = st.BboxHistory[len(st.BboxHistory)-int(historyLookbackFrames):]
	}

	headY, ok := extractHeadY(kyp)
	if ok {
		st.HeadYHistory = append(st.HeadYHistory, HeadYSnapshot{
			Timestamp: timestamp,
			Y:         headY,
		})

		if len(st.HeadYHistory) > int(historyLookbackFrames) {
			st.HeadYHistory = st.HeadYHistory[len(st.HeadYHistory)-int(historyLookbackFrames):]
		}
	}

	var validY []float64
	for i, conf := range kyp.Points.Conf {
		if float64(conf) > keypointConfThreshold && i < len(kyp.Points.XY) {
			validY = append(validY, kyp.Points.XY[i][1])
		}
	}

	bboxRatioSignal := currentRatio > ratioThresholdFallen

	keypointVarianceSignal := false
	yVariance := 0.0
	normalizedVariance := 999.0

	if len(validY) >= 5 {
		yVariance = stdDev(validY)
		normalizedVariance = yVariance / height
		keypointVarianceSignal = normalizedVariance < normalizedThreshold
	}

	velocity := calculateHeadYVelocity(st.HeadYHistory)
	velocitySignal := velocity > velocityThreshold

	signals := 0
	if bboxRatioSignal {
		signals++
	}
	if keypointVarianceSignal {
		signals++
	}
	if velocitySignal {
		signals++
	}

	fallenStateNow := signals >= 2

	log.Printf(
		"... ratio=%.2f yVar=%.2f normVar=%.2f velocity=%.2f signals=%d bboxSignal=%v keypointSignal=%v velocitySignal=%v fallenStateNow=%v ...",
		currentRatio,
		yVariance,
		normalizedVariance,
		velocity,
		signals,
		bboxRatioSignal,
		keypointVarianceSignal,
		velocitySignal,
		fallenStateNow,
	)

	if fallenStateNow {
		st.ConfirmationCount++
	} else {
		st.ConfirmationCount = 0
	}

	if st.ConfirmationCount >= confirmationFrames {
		if st.IsFallen {
			return nil
		}

		if timestamp.Sub(st.lastAlerted) < cooldownSeconds {
			return nil
		}

		st.IsFallen = true
		st.lastAlerted = timestamp

		return &Alert{
			AlertType:  "falling",
			TrackerID:  kyp.TrackerID,
			Camera:     camera,
			BoundBox:   bbox,
			RecordedAt: timestamp,
			Severity:   "critical",
			Label:      "person",
		}
	}

	if currentRatio < ratioThresholdStanding {
		st.IsFallen = false
	}

	return nil
}

func extractHeadY(kyp Keypoints) (float64, bool) {
	if len(kyp.Points.Conf) > 0 &&
		len(kyp.Points.XY) > 0 &&
		float64(kyp.Points.Conf[0]) > keypointConfThreshold {
		return kyp.Points.XY[0][1], true
	}

	leftShoulder := 5
	rightShoulder := 6

	if len(kyp.Points.Conf) > rightShoulder &&
		len(kyp.Points.XY) > rightShoulder &&
		float64(kyp.Points.Conf[leftShoulder]) > keypointConfThreshold &&
		float64(kyp.Points.Conf[rightShoulder]) > keypointConfThreshold {
		y := (kyp.Points.XY[leftShoulder][1] + kyp.Points.XY[rightShoulder][1]) / 2
		return y, true
	}

	log.Printf("[HEAD DEBUG] nose_conf=%.2f, shoulder_l_conf=%.2f, shoulder_r_conf=%.2f, returning_ok=%v", 
    kyp.Points.Conf[0], kyp.Points.Conf[5], kyp.Points.Conf[6])

	return 0, false
}

func calculateHeadYVelocity(history []HeadYSnapshot) float64 {
	if len(history) < 2 {
		return 0
	}

	oldest := history[0]
	newest := history[len(history)-1]

	timeSpan := newest.Timestamp.Sub(oldest.Timestamp).Seconds()
	if timeSpan <= 0 {
		return 0
	}

	yChange := newest.Y - oldest.Y
	return yChange / timeSpan
}

func stdDev(values []float64) float64 {

	var sum, mean float64
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	var sqSum float64
	for _, v := range values {
		sqSum += (v - mean) * (v - mean)
	}
	return math.Sqrt(sqSum / float64(len(values)))
}

func (d *FallDetector) cleanupOldStates(now time.Time, maxAge time.Duration) {
	for camera, trackers := range d.States {
		for trackerID, state := range trackers {
			if now.Sub(state.LastSeen) > maxAge {
				delete(trackers, trackerID)
			}
		}

		if len(trackers) == 0 {
			delete(d.States, camera)
		}
	}
}
