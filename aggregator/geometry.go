package main

import (
	"time"
)

func PointinPoligon(p Point, polygon []Point) bool {

	n := len(polygon)
	if n < 3 {
		return false
	}

	inside := false
	j := n - 1

	for i := 0; i < n; i++ {
		xi, yi := polygon[i].X, polygon[i].Y
		xj, yj := polygon[j].X, polygon[j].Y

		intersects := ((yi > p.Y) != (yj > p.Y)) &&
			(p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi)

		if intersects {
			inside = !inside
		}

		j = i
	}

	return inside
}

func isZoneActive(zone Zone) bool {
	if zone.ActiveFrom == nil || zone.ActiveUntil == nil {
		return true
	}

	t := time.Now()
	now := time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
	from := time.Date(0, 1, 1, zone.ActiveFrom.Hour(), zone.ActiveFrom.Minute(), zone.ActiveFrom.Second(), 0, time.UTC)
	until := time.Date(0, 1, 1, zone.ActiveUntil.Hour(), zone.ActiveUntil.Minute(), zone.ActiveUntil.Second(), 0, time.UTC)

	if !from.After(until) {
		return !now.Before(from) && !now.After(until)
	}
	return !now.Before(from) || !now.After(until)
}
