package main

import (
	"fmt"
)

// queryInfluxFloat runs a Flux query to compute an aggregate over the last `minutes` minutes.
func queryInflux(bucket, measurement, camera, field, window, agg string) string {
	return fmt.Sprintf(`
    from(bucket: %q)
      |> range(start: -%s)
      |> filter(fn: (r) =>
          r._measurement == %q and
          r.camera       == %q and
          r._field       == %q)
      |> aggregateWindow(every: %s, fn: %s)
      |> last()
  `, bucket, window, measurement, camera, field, window, agg)
}


func queryMeasurements(bucket string)string{
	return  fmt.Sprintf(`
		import "influxdata/influxdb/schema"
           schema.measurements(bucket:"` + bucket + `")
	
	`)
}

func queryCamera(bucket, measurement string)string{
	return  fmt.Sprintf(`
		from(bucket: "%s")
		  |> range(start: -1h)
		  |> filter(fn: (r) => r._measurement == "%s")
		  |> keep(columns: ["camera"])
		  |> distinct(column: "camera")
		`, bucket, measurement)
}
