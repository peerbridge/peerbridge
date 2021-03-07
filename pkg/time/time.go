package time

import (
	"fmt"
	"strings"
	"time"
)

// The default time format used to serialize and deserialize
// timestamps used in the blockchain transactions.
const ISO_8601 = time.RFC3339

// A time object which conforms to the ISO8601 specification.
type Time struct {
	time.Time
}

// Serialize a time object to JSON conforming to the ISO8601 specification.
func (t Time) MarshalJSON() ([]byte, error) {
	tstamp := fmt.Sprintf("\"%s\"", time.Time(t.Time).Format(ISO_8601))
	return []byte(tstamp), nil
}

// Deserialize a time object from a JSON value.
// The JSON value must conform to the ISO8601 specification.
func (t Time) UnmarshalJSON(b []byte) (err error) {
	timeStr := strings.Trim(string(b), `"`)

	if timeStr == "" || timeStr == "null" {
		t.Time = time.Time{}
		return
	}

	t.Time, err = time.Parse(time.RFC3339, timeStr)
	return
}

// The current time of the day.
func Now() Time {
	return Time{time.Now()}
}
