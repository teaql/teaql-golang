package core

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeOfDayNativeWallClockRoundTrip(t *testing.T) {
	for _, text := range []string{"00:00:00", "12:34:56", "23:59:59.123456789", "01:02:03.000000001"} {
		t.Run(text, func(t *testing.T) {
			value, err := ParseTimeOfDay(text)
			if err != nil || value.String() != text {
				t.Fatalf("time parse/format: %s %v", value.String(), err)
			}
			decoded, ok := ValTimeOfDay(value).TryTimeOfDay()
			if !ok || decoded != value {
				t.Fatal("typed Value lost wall-clock precision")
			}
			stored, ok := ValText(text).TryTimeOfDay()
			if !ok || stored != value {
				t.Fatal("provider text did not hydrate typed TimeOfDay")
			}
			payload, err := json.Marshal(value)
			if err != nil || string(payload) != `"`+text+`"` {
				t.Fatalf("native text JSON: %s %v", payload, err)
			}
			var restored TimeOfDay
			if err := json.Unmarshal(payload, &restored); err != nil || restored != value {
				t.Fatalf("native JSON roundtrip: %v", err)
			}
		})
	}
	var midnight TimeOfDay
	if midnight.String() != "00:00:00" || midnight.Nanoseconds() != 0 {
		t.Fatal("zero value must be midnight")
	}
	value, err := NewTimeOfDay(23, 59, 59, 999999999)
	if err != nil || value.Nanoseconds() != int64(24*time.Hour-time.Nanosecond) {
		t.Fatal("last nanosecond of day")
	}
	canonical, err := ParseTimeOfDay("12:34:56.120000000")
	if err != nil || canonical.String() != "12:34:56.12" {
		t.Fatal("canonical fractional representation")
	}
}

func TestTimeOfDayAndTimestampRemainDistinct(t *testing.T) {
	instant := time.Date(2024, 2, 29, 12, 34, 56, 789000000, time.UTC)
	value := ValTimestamp(instant.UnixMilli())
	actual, ok := value.TryTime()
	if !ok || !actual.Equal(instant) || actual.Location() != time.UTC {
		t.Fatal("native Timestamp is not exposed as its UTC instant")
	}
	if _, ok := value.TryTimeOfDay(); ok {
		t.Fatal("an instant became a wall clock")
	}
	clock, _ := ParseTimeOfDay("12:34:56.789")
	if _, ok := ValTimeOfDay(clock).TryTime(); ok {
		t.Fatal("a wall clock acquired a date/timezone")
	}
}

func TestTimeOfDayRejectsInvalidValuesWithoutChangingTheReceiver(t *testing.T) {
	for _, text := range []string{"", "12:34", "24:00:00", "12:60:00", "12:00:60", "-1:00:00", "1:02:03", "12:00:00.",
		"12:00:00.1234567890", "12:00:00Z", "12:00:00+08:00", "2024-01-01T12:00:00Z", " 12:00:00", "12:00:00\n", "１２:00:00"} {
		t.Run(text, func(t *testing.T) {
			if _, err := ParseTimeOfDay(text); err == nil {
				t.Fatal("invalid wall clock accepted")
			}
			value, _ := ParseTimeOfDay("12:34:56")
			before := value
			if err := value.UnmarshalText([]byte(text)); err == nil || value != before {
				t.Fatal("failed parse changed existing value")
			}
		})
	}
	for _, fields := range [][4]int{{-1, 0, 0, 0}, {24, 0, 0, 0}, {0, -1, 0, 0}, {0, 60, 0, 0}, {0, 0, -1, 0}, {0, 0, 60, 0}, {0, 0, 0, -1}, {0, 0, 0, 1000000000}} {
		if _, err := NewTimeOfDay(fields[0], fields[1], fields[2], fields[3]); err == nil {
			t.Fatal("out-of-range constructor accepted")
		}
	}
	for _, value := range []Value{ValNull(), ValI64(0), ValBool(false), ValDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), ValText("not a time")} {
		if _, ok := value.TryTimeOfDay(); ok {
			t.Fatal("unrelated scalar coerced to wall clock")
		}
	}
	var value TimeOfDay
	for _, payload := range []string{`null`, `0`, `true`, `{}`, `"24:00:00"`} {
		if err := json.Unmarshal([]byte(payload), &value); err == nil {
			t.Fatal("non-wall-clock JSON accepted: " + payload)
		}
	}
}
