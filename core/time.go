package core

import "time"

// Timestamp represents a Unix timestamp in milliseconds.
type Timestamp int64

func TimestampNow() Timestamp {
	return Timestamp(time.Now().UnixMilli())
}

func TimestampFromUint64(val uint64) Timestamp {
	return Timestamp(val)
}

func (t Timestamp) AsMillis() int64 {
	return int64(t)
}

func (t Timestamp) ToTime() time.Time {
	return time.UnixMilli(int64(t)).UTC()
}
