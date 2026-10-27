package core

import (
	"errors"
	"math"
	"testing"
)

func TestVersionLifecycleCoreCheckedTransitions(t *testing.T) {
	for _, item := range []struct {
		name       string
		next       func(int64) (int64, error)
		from, want int64
	}{
		{"ordinary update", VersionAfterUpdate, 1, 2},
		{"last update", VersionAfterUpdate, math.MaxInt64 - 1, math.MaxInt64},
		{"ordinary deletion", VersionAfterSoftDelete, 1, -2},
		{"maximum deletion", VersionAfterSoftDelete, math.MaxInt64, math.MinInt64},
		{"first tombstone recovery", VersionAfterRecovery, -1, 2},
		{"ordinary recovery", VersionAfterRecovery, -2, 3},
		{"last recovery", VersionAfterRecovery, math.MinInt64 + 2, math.MaxInt64},
	} {
		t.Run(item.name, func(t *testing.T) {
			got, err := item.next(item.from)
			if err != nil || got != item.want {
				t.Fatalf("got=%d err=%v want=%d", got, err, item.want)
			}
		})
	}
}

func TestVersionLifecycleCoreTypedRejections(t *testing.T) {
	for _, item := range []struct {
		name      string
		next      func(int64) (int64, error)
		versions  []int64
		exhausted bool
	}{
		{"invalid update", VersionAfterUpdate, []int64{math.MinInt64, -1, 0}, false},
		{"exhausted update", VersionAfterUpdate, []int64{math.MaxInt64}, true},
		{"invalid deletion", VersionAfterSoftDelete, []int64{math.MinInt64, -1, 0}, false},
		{"invalid recovery", VersionAfterRecovery, []int64{0, 1, math.MaxInt64}, false},
		{"exhausted recovery", VersionAfterRecovery, []int64{math.MinInt64, math.MinInt64 + 1}, true},
	} {
		t.Run(item.name, func(t *testing.T) {
			for _, version := range item.versions {
				value, err := item.next(version)
				var typed *MutationVersionError
				if value != 0 || !errors.As(err, &typed) || typed.OriginalVersion != version || typed.Exhausted != item.exhausted {
					t.Fatalf("version=%d value=%d wrong typed error: %+v", version, value, err)
				}
			}
		})
	}
}
