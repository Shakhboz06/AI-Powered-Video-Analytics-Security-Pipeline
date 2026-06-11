package main

import (
	"log"
	"math"
	"time"
)

type AbandonedObjectDetector struct {
	States      map[string]map[int64]*PerObjectState
	LastCleanup time.Time
}

type PerObjectState struct {
	StationaryAnchor [4]float64
	StationarySince  time.Time
	isAlerted        bool
	lastAlerted      time.Time
	LastSeen         time.Time
}

func NewAbandonedObjectDetector() *AbandonedObjectDetector {
	return &AbandonedObjectDetector{
		States:      make(map[string]map[int64]*PerObjectState),
		LastCleanup: time.Now(),
	}
}

const (
	// for production stationaryThreshold is 30 seconds i need to revert it to 30 sec for dev it is enough just to check
	stationaryThreshold      time.Duration = time.Second * 5
	movementThreshold        float64       = 20
	personProximityThreshold float64       = 200
	historyLookback          time.Duration = time.Second * 5
	abandonedCooldown        time.Duration = 60 * time.Second
)

var abandonedObjectLabels = map[string]bool{
	"backpack": true,
	"handbag":  true,
	"suitcase": true,
	"luggage":  true,
}

func (d *AbandonedObjectDetector) DetectAbandonedObj(camera string, detections []Detections, timestamp time.Time) []Alert {

	if timestamp.Sub(d.LastCleanup) >= 10*time.Second {
		d.cleanupOldStates(timestamp, 60*time.Second)
		d.LastCleanup = timestamp
	}

	var alerts []Alert

	var objects []Detections
	var persons []Detections

	for _, det := range detections {

		if det.Label == "person" {
			persons = append(persons, det)
		} else if abandonedObjectLabels[det.Label] {
			objects = append(objects, det)
		}
	}

	for _, obj := range objects {

		if d.States[camera] == nil {
			d.States[camera] = make(map[int64]*PerObjectState)
		}

		state, exists := d.States[camera][obj.TrackerID]

		if !exists {
			state = &PerObjectState{
				StationarySince:  timestamp,
				StationaryAnchor: obj.BoundBox,
			}
			d.States[camera][obj.TrackerID] = state

		}

		state.LastSeen = timestamp

		anchorCenterX := (state.StationaryAnchor[0] + state.StationaryAnchor[2]) / 2
		anchorCenterY := (state.StationaryAnchor[1] + state.StationaryAnchor[3]) / 2

		currentCenterX := (obj.BoundBox[0] + obj.BoundBox[2]) / 2
		currentCenterY := (obj.BoundBox[1] + obj.BoundBox[3]) / 2

		dx := currentCenterX - anchorCenterX
		dy := currentCenterY - anchorCenterY
		distFromAnchor := math.Sqrt(dx*dx + dy*dy)

		if distFromAnchor > movementThreshold {
			state.StationarySince = timestamp
			state.StationaryAnchor = obj.BoundBox
			state.isAlerted = false
			continue
		}

		log.Printf("[ABANDONED DEBUG] tracker=%d label=%s dist_from_anchor=%.1f stationary_duration=%.1fs num_persons=%d",
			obj.TrackerID,
			obj.Label,
			distFromAnchor,
			timestamp.Sub(state.StationarySince).Seconds(),
			len(persons),
		)

		if timestamp.Sub(state.StationarySince) < stationaryThreshold {
			continue
		}
		
		log.Printf("[ABANDONED DEBUG] tracker=%d passed duration check, num_persons=%d", obj.TrackerID, len(persons))

		personNearby := false
		for _, p := range persons {

			personCenterX := (p.BoundBox[0] + p.BoundBox[2]) / 2
			personCenterY := (p.BoundBox[1] + p.BoundBox[3]) / 2

			dx := currentCenterX - personCenterX
			dy := currentCenterY - personCenterY
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance < personProximityThreshold {
				personNearby = true
				break
			}
		}
		log.Printf("[ABANDONED DEBUG] tracker=%d person_nearby=%v stationary_for=%.1fs",
			obj.TrackerID,
			personNearby,
			timestamp.Sub(state.StationarySince).Seconds(),
		)
		if personNearby {
			continue
		}

		if state.isAlerted {
			continue
		}

		if timestamp.Sub(state.lastAlerted) < abandonedCooldown {
			continue
		}

		state.isAlerted = true
		state.lastAlerted = timestamp

		alerts = append(alerts, Alert{
			Camera:     camera,
			TrackerID:  obj.TrackerID,
			BoundBox:   obj.BoundBox,
			Label:      obj.Label,
			RecordedAt: timestamp,
			Severity:   "high",
			AlertType:  "abandoned_object",
		})
	}

	return alerts
}

func (d *AbandonedObjectDetector) cleanupOldStates(now time.Time, maxAge time.Duration) {
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
