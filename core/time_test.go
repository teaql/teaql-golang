package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampConversions(t *testing.T) {
	ts := Timestamp(1000)
	assert.Equal(t, int64(1000), ts.AsMillis())
	dt := ts.ToTime()
	assert.Equal(t, int64(1000), dt.UnixMilli())

	ts2 := TimestampFromUint64(2000)
	assert.Equal(t, int64(2000), ts2.AsMillis())
}

func TestTimestampNow(t *testing.T) {
	before := time.Now().UTC().UnixMilli()
	ts := TimestampNow()
	after := time.Now().UTC().UnixMilli()
	
	assert.GreaterOrEqual(t, ts.AsMillis(), before)
	assert.LessOrEqual(t, ts.AsMillis(), after)
}
