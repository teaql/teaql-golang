package runtime

import "time"

// BusinessClock supplies the time used by domain behavior, Checker/Fix, and
// other business decisions. Operational concerns such as cache expiry and
// telemetry durations must continue to use their own monotonic/system clocks.
type BusinessClock interface {
	Now() time.Time
}

// SystemBusinessClock is the default business clock.
type SystemBusinessClock struct{}

func (SystemBusinessClock) Now() time.Time { return time.Now() }

// FixedBusinessClock is a deterministic clock intended for tests and replay.
type FixedBusinessClock struct {
	value time.Time
}

func NewFixedBusinessClock(value time.Time) FixedBusinessClock {
	return FixedBusinessClock{value: value}
}

func (c FixedBusinessClock) Now() time.Time { return c.value }
