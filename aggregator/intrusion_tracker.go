package main

import (
	"time"
)

type ZoneTrackerState struct {
	inZone map[int64]map[int64]TrackerRecorder
}

type TrackerRecorder struct {
	EntryTime     time.Time
	MissedFrames  int
	LoiterAlerted bool
}

func NewZoneTrackerState() *ZoneTrackerState {
	return &ZoneTrackerState{
		inZone: make(map[int64]map[int64]TrackerRecorder),
	}
}

type Alert struct {
	Camera     string     `json:"camera"`
	ZoneID     *int64     `json:"zone_id"`
	ZoneName   string     `json:"zone_name,omitempty"`
	TrackerID  int64      `json:"tracker_id"`
	BoundBox   [4]float64 `json:"bound_box"`
	Label      string     `json:"label"`
	RecordedAt time.Time  `json:"recorded_at"`
	AlertType  string     `json:"alert_type"`
	Severity   string     `json:"severity"`
}

const (
	GRACE_PERIOD       = 5
	AlertTypeIntrusion = "intrusion"
	AlertTypeLoitering = "loitering"
)

func (s *ZoneTrackerState) UpdateState(current map[int64]map[int64]Alert, zonesByID map[int64]Zone, now time.Time) []Alert {
	
	var alerts []Alert
	newState := map[int64]map[int64]TrackerRecorder{}

	for zoneID, currentTracker := range current {

		previuosTracker := s.inZone[zoneID]
		var newAdd TrackerRecorder

		for trackerID, item := range currentTracker {

			if newState[zoneID] == nil {
				newState[zoneID] = map[int64]TrackerRecorder{}
			}

			prev, wasIn := previuosTracker[trackerID]

			if !wasIn {
				alerts = append(alerts, Alert{
					ZoneID:     &zoneID,
					ZoneName:   item.ZoneName,
					TrackerID:  trackerID,
					Camera:     item.Camera,
					BoundBox:   item.BoundBox,
					Label:      item.Label,
					RecordedAt: item.RecordedAt,
					AlertType:  AlertTypeIntrusion,
					Severity:   zonesByID[zoneID].DefaultSeverity,
				})
				newAdd = TrackerRecorder{
					EntryTime:     now,
					MissedFrames:  0,
					LoiterAlerted: false}
			} else {
				newAdd = prev
				newAdd.MissedFrames = 0
			}

			if zonesByID[zoneID].LoiterThreshold != nil {
				if *zonesByID[zoneID].LoiterThreshold > 0 && !newAdd.LoiterAlerted {
					duration := now.Sub(newAdd.EntryTime)
					if duration > time.Duration(*zonesByID[zoneID].LoiterThreshold)*time.Second {
						alerts = append(alerts, Alert{
							ZoneID:     &zoneID,
							ZoneName:   item.ZoneName,
							TrackerID:  trackerID,
							Camera:     item.Camera,
							BoundBox:   item.BoundBox,
							Label:      item.Label,
							RecordedAt: item.RecordedAt,
							AlertType:  AlertTypeLoitering,
							Severity:   "warning",
						})
						newAdd.LoiterAlerted = true
					}
				}
			}
			newState[zoneID][trackerID] = newAdd
		}
	}

	for zoneID, prev := range s.inZone {
		for trackerID, exist := range prev {

			if newState[zoneID] == nil {
				newState[zoneID] = map[int64]TrackerRecorder{}
			}
			if _, present := current[zoneID][trackerID]; present {
				continue
			}

			exist.MissedFrames++
			if exist.MissedFrames <= GRACE_PERIOD {
				newState[zoneID][trackerID] = exist
			}
		}
	}

	s.inZone = newState

	return alerts
}
