package runtime

import (
	"math"
	"testing"
	"time"
)

func TestMemoryDataStoreLifetimeBounds(t *testing.T) {
	const maxSeconds = uint64((1<<63 - 1) / int64(time.Second))
	zero, one, maximum, over, huge := uint64(0), uint64(1), maxSeconds, maxSeconds+1, uint64(math.MaxUint64)
	for _, test := range []struct {
		name    string
		seconds *uint64
		want    time.Duration
	}{
		{"nil", nil, 0}, {"zero", &zero, 0}, {"one", &one, time.Second},
		{"maximum", &maximum, time.Duration(maxSeconds) * time.Second},
		{"over-maximum", &over, time.Duration(maxSeconds) * time.Second},
		{"unsigned-maximum", &huge, time.Duration(maxSeconds) * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			if actual := memoryDataStoreLifetime(test.seconds); actual != test.want {
				t.Fatalf("wrong bounded lifetime: got %v want %v", actual, test.want)
			}
		})
	}
}
