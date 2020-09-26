package time

import (
	"fmt"
	"strings"
	"time"
)

const ISO_8601 = time.RFC3339

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	tstamp := fmt.Sprintf("\"%s\"", time.Time(t.Time).Format(ISO_8601))
	return []byte(tstamp), nil
}

func (t Time) UnmarshalJSON(b []byte) (err error) {
	timeStr := strings.Trim(string(b), `"`)

	if timeStr == "" || timeStr == "null" {
		t.Time = time.Time{}
		return
	}

	t.Time, err = time.Parse(time.RFC3339, timeStr)
	return
}

func Now() Time {
	return Time{time.Now()}
}
