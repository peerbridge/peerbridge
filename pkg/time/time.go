package time

import (
	"fmt"
	"time"
)

const ISO_8601 = time.RFC3339

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	tstamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(ISO_8601))
	return []byte(tstamp), nil
}

func Now() Time {
	return Time(time.Now())
}
