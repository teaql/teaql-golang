package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TimeOfDay is a wall clock, not an instant or an elapsed duration. Its zero
// value is midnight. Constructors are the only way to create a nonzero value.
// It carries neither a calendar date, timezone, context nor system clock.
type TimeOfDay struct{ nanoseconds int64 }

var ErrInvalidTimeOfDay = errors.New("invalid TimeOfDay: expected HH:mm:ss with at most nine fractional digits")

func NewTimeOfDay(hour, minute, second, nanosecond int) (TimeOfDay, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 || nanosecond < 0 || nanosecond >= 1000000000 {
		return TimeOfDay{}, ErrInvalidTimeOfDay
	}
	return TimeOfDay{int64(hour)*int64(time.Hour) + int64(minute)*int64(time.Minute) + int64(second)*int64(time.Second) + int64(nanosecond)}, nil
}

func ParseTimeOfDay(text string) (TimeOfDay, error) {
	if len(text) < 8 || len(text) > 18 || text[2] != ':' || text[5] != ':' {
		return TimeOfDay{}, ErrInvalidTimeOfDay
	}
	for _, position := range []int{0, 1, 3, 4, 6, 7} {
		if text[position] < '0' || text[position] > '9' {
			return TimeOfDay{}, ErrInvalidTimeOfDay
		}
	}
	nanos := 0
	if len(text) > 8 {
		if text[8] != '.' || len(text) == 9 {
			return TimeOfDay{}, ErrInvalidTimeOfDay
		}
		for index := 9; index < len(text); index++ {
			if text[index] < '0' || text[index] > '9' {
				return TimeOfDay{}, ErrInvalidTimeOfDay
			}
			nanos = nanos*10 + int(text[index]-'0')
		}
		for index := len(text) - 9; index < 9; index++ {
			nanos *= 10
		}
	}
	return NewTimeOfDay(int(text[0]-'0')*10+int(text[1]-'0'), int(text[3]-'0')*10+int(text[4]-'0'), int(text[6]-'0')*10+int(text[7]-'0'), nanos)
}

func (value TimeOfDay) Nanoseconds() int64 { return value.nanoseconds }

func (value TimeOfDay) String() string {
	seconds := value.nanoseconds / int64(time.Second)
	text := fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)
	if fraction := value.nanoseconds % int64(time.Second); fraction != 0 {
		text += "." + strings.TrimRight(fmt.Sprintf("%09d", fraction), "0")
	}
	return text
}

func (value TimeOfDay) MarshalText() ([]byte, error) { return []byte(value.String()), nil }

func (value *TimeOfDay) UnmarshalText(text []byte) error {
	parsed, err := ParseTimeOfDay(string(text))
	if err != nil {
		return err
	}
	if value == nil {
		return ErrInvalidTimeOfDay
	}
	*value = parsed
	return nil
}

// encoding/json otherwise ignores null for TextUnmarshaler structs. Reject it
// here; an optional field is represented by *TimeOfDay instead.
func (value *TimeOfDay) UnmarshalJSON(payload []byte) error {
	var text string
	if err := json.Unmarshal(payload, &text); err != nil {
		return ErrInvalidTimeOfDay
	}
	return value.UnmarshalText([]byte(text))
}

func ValTimeOfDay(value TimeOfDay) Value { return Value{value} }

// Providers may return SQL TIME as text. Never interpret a date, timestamp,
// numeric duration or timezone-bearing instant as a wall-clock value.
func (value Value) TryTimeOfDay() (TimeOfDay, bool) {
	switch typed := value.V.(type) {
	case TimeOfDay:
		return typed, true
	case string:
		parsed, err := ParseTimeOfDay(typed)
		return parsed, err == nil
	default:
		return TimeOfDay{}, false
	}
}
