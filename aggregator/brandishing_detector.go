package main

import (
	"log"
	"math"
	"time"
)

type PerPersonBrandishingState struct {
	ConfirmationCount int64
	IsBrandishing     bool
	lastAlerted       time.Time
	LastSeen          time.Time
}

type BrandishingDetector struct {
	States      map[string]map[int64]*PerPersonBrandishingState
	LastCleanup time.Time
}

const (
	brandishingDistanceThreshold float64       = 80
	brandishingConfirmFrames     int64         = 3
	brandishingCooldown          time.Duration = 30 * time.Second
)

var weaponLabels = map[string]bool{
	"rifle":  true,
	"pistol": true,
	"knife":  true,
}

func NewBrandishingDetector() *BrandishingDetector {
	return &BrandishingDetector{
		States:      make(map[string]map[int64]*PerPersonBrandishingState),
		LastCleanup: time.Now(),
	}
}

func (d *BrandishingDetector) DetectBrandishing(camera string, detections []Detections, keypoints []Keypoints, timestamp time.Time) []Alert {

	if timestamp.Sub(d.LastCleanup) >= 10*time.Second {
		d.cleanupOldStates(timestamp, 60*time.Second)
		d.LastCleanup = timestamp
	}

	var alerts []Alert
	keypointMap := make(map[int64]Keypoints)

	for _, kp := range keypoints {
		keypointMap[kp.TrackerID] = kp
	}

	var weapons []Detections

	for _, det := range detections {
		if weaponLabels[det.Label] {
			weapons = append(weapons, det)
		}

	}

	for _, det := range detections {

		if det.Label != "person" {
			continue
		}

		kyp, exist := keypointMap[det.TrackerID]
		if !exist {
			continue
		}

		if d.States[camera] == nil {
			d.States[camera] = make(map[int64]*PerPersonBrandishingState)
		}

		state, exists := d.States[camera][det.TrackerID]
		if !exists {
			state = &PerPersonBrandishingState{}
			d.States[camera][det.TrackerID] = state
		}
		state.LastSeen = timestamp

		left_w := kyp.Points.XY[9]
		right_w := kyp.Points.XY[10]
		left_conf := kyp.Points.Conf[9]
		right_conf := kyp.Points.Conf[10]

		near_weapon := false

		for _, weapon := range weapons {

			centerX := (weapon.BoundBox[0] + weapon.BoundBox[2]) / 2
			centerY := (weapon.BoundBox[1] + weapon.BoundBox[3]) / 2

			if left_conf > keypointConfThreshold {
				if distance(centerX, centerY, left_w[0], left_w[1]) < brandishingDistanceThreshold {
					near_weapon = true
					break
				}
			}

			if right_conf > keypointConfThreshold {
				if distance(centerX, centerY, right_w[0], right_w[1]) < brandishingDistanceThreshold {
					near_weapon = true
					break
				}
			}
		}

		if near_weapon {
			state.ConfirmationCount++
		} else {
			state.ConfirmationCount = 0
			state.IsBrandishing = false
		}

		log.Printf("tracker=%d num_weapons=%d near=%v conf_count=%d",
			det.TrackerID, len(weapons), near_weapon, state.ConfirmationCount)
		if state.ConfirmationCount >= brandishingConfirmFrames {
			if timestamp.Sub(state.lastAlerted) < brandishingCooldown {
				continue
			}

			if state.IsBrandishing {
				continue
			}

			state.IsBrandishing = true
			state.lastAlerted = timestamp

			alerts = append(alerts, Alert{
				Camera:     camera,
				TrackerID:  kyp.TrackerID,
				BoundBox:   det.BoundBox,
				Label:      det.Label,
				RecordedAt: timestamp,
				Severity:   "critical",
				AlertType:  "brandishing",
			})
		}
	}

	return alerts
}

func distance(p1x, p1y, p2x, p2y float64) float64 {
	dx := p1x - p2x
	dy := p1y - p2y
	return math.Sqrt(dx*dx + dy*dy)
}

func (d *BrandishingDetector) cleanupOldStates(now time.Time, maxAge time.Duration) {
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
