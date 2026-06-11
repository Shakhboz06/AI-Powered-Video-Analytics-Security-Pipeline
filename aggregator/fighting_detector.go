package main

import (
	"time"
)

type FightDetector struct {
	cooldowns         map[string]time.Time
	predictionHistory map[string][]float64
}

func NewFightDetector() *FightDetector {
	return &FightDetector{
		cooldowns:         make(map[string]time.Time),
		predictionHistory: make(map[string][]float64),
	}
}

const (
	fightCooldownDuration  = 30 * time.Second
	fightThreshold         = 0.6
	maxHistorySize         = 5
	minHistoryForInference = 3
)

func (d *FightDetector) DetectFight(camera string, detections []Detections, prediction *float64, timestamp time.Time) *Alert {

	if prediction == nil {
		return nil
	}

	var people []Detections

	for _, det := range detections {
		if det.Label == "person" {
			people = append(people, det)
		}
	}


	d.predictionHistory[camera] = append(d.predictionHistory[camera], *prediction)

	if len(d.predictionHistory[camera]) > maxHistorySize {
		d.predictionHistory[camera] = d.predictionHistory[camera][len(d.predictionHistory[camera])-maxHistorySize:]
	}

	if len(d.predictionHistory[camera]) < minHistoryForInference{
		return nil
	}

	sum := 0.0

	for _, p := range d.predictionHistory[camera]{
		sum += p
	}

	avg := sum / float64(len(d.predictionHistory[camera]))

	if avg <= fightThreshold{
		return nil
	}

	if d.fightCooldownActive(camera, timestamp) {
		return nil
	}

	if len(people) < 2 {
		return nil
	}

	left := people[0].BoundBox[0]
	top := people[0].BoundBox[1]
	right := people[0].BoundBox[2]
	bottom := people[0].BoundBox[3]

	for _, p := range people[1:] {
		if p.BoundBox[0] < left {
			left = p.BoundBox[0]
		}

		if p.BoundBox[1] < top {
			top = p.BoundBox[1]
		}

		if p.BoundBox[2] > right {
			right = p.BoundBox[2]
		}

		if p.BoundBox[3] > bottom {
			bottom = p.BoundBox[3]
		}
	}

	groupBox := [4]float64{left, top, right, bottom}

	alert := Alert{
		AlertType:  "fighting",
		Camera:     camera,
		Severity:   "high",
		RecordedAt: timestamp,
		Label:      "person_group",
		BoundBox:   groupBox,
		TrackerID:  0,
	}

	d.cooldowns[camera] = timestamp

	return &alert

}

func (d *FightDetector) fightCooldownActive(camera string, now time.Time) bool {
	lastAlerted, exists := d.cooldowns[camera]
	if !exists {
		return false
	}

	return now.Sub(lastAlerted) < fightCooldownDuration
}
