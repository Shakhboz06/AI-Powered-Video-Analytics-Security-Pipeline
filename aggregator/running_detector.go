package main

import (
	"math"
	"time"
)

type PositionEntry struct {
	Position  Point
	Timestamp time.Time
}

type RunningDetector struct {
	history     map[int64][]PositionEntry
	lastAlerted map[int64]time.Time
}

func NewRunningDetector() *RunningDetector {
	return &RunningDetector{
		history:     make(map[int64][]PositionEntry),
		lastAlerted: make(map[int64]time.Time),
	}
}

const (
	AlertTypeRunning = "running"
	RUNNING_PIXEL_THRESHOLD  = 100
	RUNNING_HISTORY_DURATION = time.Second
	RUNNING_ALERT_COOLDOWN   = time.Second * 5
)

func (d *RunningDetector) DetectRunning(camera string, detections []Detections, now time.Time) []Alert {

	var posEntry PositionEntry
	var alert [] Alert
	for _, item := range detections {

		x_min := item.BoundBox[0]
		x_max := item.BoundBox[2]
		y_max := item.BoundBox[3]

		point := Point{
			X: (x_min + x_max) / 2,
			Y: y_max,
		}

		posEntry = PositionEntry{
			Position:  point,
			Timestamp: now,
		}

		d.history[item.TrackerID] = append(d.history[item.TrackerID], posEntry)

		var pruned []PositionEntry
		for _, t := range d.history[item.TrackerID] {

			if t.Timestamp.After(now.Add(-RUNNING_HISTORY_DURATION)) {
				pruned = append(pruned, t)
			}

		}

		d.history[item.TrackerID] = pruned

		if len(d.history[item.TrackerID]) < 2 {
			continue
		}

		last := d.history[item.TrackerID][len(d.history[item.TrackerID])-1]
		first := d.history[item.TrackerID][0]

		dx := last.Position.X - first.Position.X
		dy := last.Position.Y - first.Position.Y
		dt := last.Timestamp.Sub(first.Timestamp).Seconds()

		vel := math.Sqrt(dx*dx+dy*dy) / dt

		if vel > RUNNING_PIXEL_THRESHOLD {
			if now.Sub(d.lastAlerted[item.TrackerID]) > RUNNING_ALERT_COOLDOWN{
				alert = append(alert, Alert{
					Camera: camera,
					TrackerID: item.TrackerID,
					ZoneID: nil,
					BoundBox: item.BoundBox,
					Label: item.Label,
					RecordedAt: now,
					AlertType: AlertTypeRunning,
					Severity: "critical",
				})

				d.lastAlerted[item.TrackerID] = now
			}
		}
	}

	return alert
}
