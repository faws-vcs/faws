package timestamp

import (
	"strconv"
	"time"
)

var layout = "02.01.2006"

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

func Format(ts int64) string {
	return time.Unix(ts, 0).UTC().Format(layout)
}
