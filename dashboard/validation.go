package main

import "fmt"

var (
	allowedMeasurements = map[string]struct{}{
		"frame_detections": {},
		"frame_enriched":   {},
	}
	allowedFields = map[string]struct{}{
		"cars":       {},
		"persons":    {},
		"latency_ms": {},
	}
	allowedAggFns = map[string]struct{}{
		"mean": {},
		"max":  {},
		"min":  {},
		"sum":  {},
	}
)

func validateParam(name, value string, allowed map[string]struct{}) error {
	if _, ok := allowed[value]; !ok {
		return fmt.Errorf("invalid %s: %q", name, value)
	}
	return nil
}
