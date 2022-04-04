package anti_entropy

import (
	"strconv"
	"time"
)

// MarshalJSON is used to convert the timestamp to JSON
func (t EpochTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *EpochTime) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t EpochTime) Unix() int64 {
	return time.Time(t).Unix()
}

// This returns the Unix time in nanoseconds.
func (t EpochTime) UnixNano() int64 {
	return time.Time(t).UnixNano()
}

// Time returns the JSON time as a time.Time instance in UTC
func (t EpochTime) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t EpochTime) String() string {
	return t.Time().String()
}
