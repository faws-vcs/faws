package timestamp

import (
	"strconv"
	"time"
)

var layout = "02.01.2006"

// Parse converts either:
// 1. a unix seconds timestamp string
// 2. a DD/MM/YYYY date string
// into an int64 representing unix seconds
func Parse(str string) (ts int64, err error) {
	ts, err = strconv.ParseInt(str, 10, 64)
	if err == nil {
		return
	}

	var t time.Time
	t, err = time.Parse(layout, str)
	if err != nil {
		return
	}

	ts = t.Unix()
	return
}

// Format converts a unix seconds timestamp into a DD/MM/YYYY date string
func Format(ts int64) string {
	return time.Unix(ts, 0).UTC().Format(layout)
}
