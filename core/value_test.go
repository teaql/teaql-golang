package core

import (
	"math"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestValue_TryI64_AcceptsRepresentableNumericVariants(t *testing.T) {
	assertTryI64(t, ValI64(math.MinInt64), math.MinInt64, true)
	assertTryI64(t, ValI64(math.MaxInt64), math.MaxInt64, true)
	assertTryI64(t, ValU64(uint64(math.MaxInt64)), math.MaxInt64, true)
	assertTryI64(t, ValDecimal(decimal.NewFromInt(-42)), -42, true)
}

func TestValue_TryI64_RejectsUnsignedOverflowAndUnrelatedVariants(t *testing.T) {
	assertTryI64(t, ValU64(uint64(math.MaxInt64)+1), 0, false)
	assertTryI64(t, ValU64(math.MaxUint64), 0, false)
	assertTryI64(t, ValF64(42.0), 0, false)
	assertTryI64(t, ValText("42"), 0, false)
	assertTryI64(t, ValNull(), 0, false)
}

func TestValue_TryU64_AcceptsRepresentableNumericVariants(t *testing.T) {
	assertTryU64(t, ValU64(0), 0, true)
	assertTryU64(t, ValU64(math.MaxUint64), math.MaxUint64, true)
	assertTryU64(t, ValI64(math.MaxInt64), uint64(math.MaxInt64), true)
	assertTryU64(t, ValDecimal(decimal.NewFromInt(42)), 42, true)
}

func TestValue_TryU64_RejectsNegativeAndUnrelatedVariants(t *testing.T) {
	assertTryU64(t, ValI64(-1), 0, false)
	assertTryU64(t, ValDecimal(decimal.NewFromInt(-1)), 0, false)
	assertTryU64(t, ValF64(42.0), 0, false)
	assertTryU64(t, ValText("42"), 0, false)
	assertTryU64(t, ValNull(), 0, false)
}

func TestValue_TryDecimal_AcceptsDecimalIntegerAndTextVariants(t *testing.T) {
	d, _ := decimal.NewFromString("123.450")
	assertTryDecimal(t, ValDecimal(d), d, true)
	assertTryDecimal(t, ValI64(math.MinInt64), decimal.NewFromInt(math.MinInt64), true)
	// uint64 max is tricky for decimal in some cases, but decimal handles it via string or big int
	uMax := decimal.NewFromUint64(math.MaxUint64)
	assertTryDecimal(t, ValU64(math.MaxUint64), uMax, true)
	assertTryDecimal(t, ValText("123.450"), d, true)
	assertTryDecimal(t, ValF64(136.25), decimal.NewFromFloat(136.25), true)
}

func TestValue_TryDecimal_RejectsInvalidTextAndUnrelatedVariants(t *testing.T) {
	assertTryDecimal(t, ValText("not-a-decimal"), decimal.Zero, false)
	assertTryDecimal(t, ValBool(true), decimal.Zero, false)
	assertTryDecimal(t, ValF64(math.NaN()), decimal.Zero, false)
	assertTryDecimal(t, ValNull(), decimal.Zero, false)
}

func TestValue_TryF64_AcceptsSupportedNumericVariants(t *testing.T) {
	assertTryF64(t, ValF64(1.25), 1.25, true)
	assertTryF64(t, ValI64(-2), -2.0, true)
	assertTryF64(t, ValU64(2), 2.0, true)
	d, _ := decimal.NewFromString("1.5")
	assertTryF64(t, ValDecimal(d), 1.5, true)
}

func TestValue_TryF64_RejectsUnrelatedVariants(t *testing.T) {
	assertTryF64(t, ValText("1.5"), 0, false)
	assertTryF64(t, ValBool(true), 0, false)
	assertTryF64(t, ValNull(), 0, false)
}

func TestValue_TryDate_AcceptsDateAndIsoDateText(t *testing.T) {
	leapDay := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	assertTryDate(t, ValDate(leapDay), leapDay, true)
	assertTryDate(t, ValText("2024-02-29"), leapDay, true)

	millis := leapDay.UnixMilli()
	assertTryDate(t, ValI64(millis), leapDay, true)
	assertTryDate(t, ValU64(uint64(millis)), leapDay, true)
}

func TestValue_TryDate_RejectsInvalidDatesAndUnrelatedVariants(t *testing.T) {
	assertTryDate(t, ValText("2023-02-29"), time.Time{}, false)
	assertTryDate(t, ValText("2024-02-29T00:00:00Z"), time.Time{}, false)
	assertTryDate(t, ValNull(), time.Time{}, false)
}

func TestValue_TryTimestamp_AcceptsTimestampAndSupportedTextFormats(t *testing.T) {
	utcTime, _ := time.Parse(time.RFC3339, "2024-01-02T03:04:05Z")
	utcTimestamp := utcTime.UnixMilli()

	offsetTime, _ := time.Parse(time.RFC3339, "2024-01-02T03:04:05+08:00")
	offsetTimestamp := offsetTime.UTC().UnixMilli()

	// "2024-01-02 03:04:05" parsed as UTC
	naiveTime, _ := time.Parse("2006-01-02 15:04:05", "2024-01-02 03:04:05")
	naiveTimestamp := naiveTime.UnixMilli()

	// "2024-01-02" parsed as UTC midnight
	midnightTime, _ := time.Parse("2006-01-02", "2024-01-02")
	midnightTimestamp := midnightTime.UnixMilli()

	assertTryTimestamp(t, ValTimestamp(utcTimestamp), utcTimestamp, true)
	assertTryTimestamp(t, ValText("2024-01-02T03:04:05+08:00"), offsetTimestamp, true)
	assertTryTimestamp(t, ValText("2024-01-02 03:04:05"), naiveTimestamp, true)
	assertTryTimestamp(t, ValText("2024-01-02"), midnightTimestamp, true)

	assertTryTimestamp(t, ValI64(utcTimestamp), utcTimestamp, true)
	assertTryTimestamp(t, ValU64(uint64(utcTimestamp)), utcTimestamp, true)
}

func TestValue_TryTimestamp_NormalizesOffsetsAndRejectsInvalidInput(t *testing.T) {
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-01T19:04:05Z")
	expectedUtc := expectedTime.UnixMilli()

	assertTryTimestamp(t, ValText("2024-01-02T03:04:05+08:00"), expectedUtc, true)
	assertTryTimestamp(t, ValText("2024-13-40 25:61:61"), 0, false)
	assertTryTimestamp(t, ValBool(true), 0, false)
	assertTryTimestamp(t, ValNull(), 0, false)
}

// --- Test Helpers ---

func assertTryI64(t *testing.T, v Value, expected int64, ok bool) {
	t.Helper()
	val, okRes := v.TryI64()
	assert.Equal(t, ok, okRes)
	if ok {
		assert.Equal(t, expected, val)
	}
}

func assertTryU64(t *testing.T, v Value, expected uint64, ok bool) {
	t.Helper()
	val, okRes := v.TryU64()
	assert.Equal(t, ok, okRes)
	if ok {
		assert.Equal(t, expected, val)
	}
}

func assertTryDecimal(t *testing.T, v Value, expected decimal.Decimal, ok bool) {
	t.Helper()
	val, okRes := v.TryDecimal()
	assert.Equal(t, ok, okRes)
	if ok {
		assert.True(t, expected.Equal(val), "expected %v got %v", expected, val)
	}
}

func assertTryF64(t *testing.T, v Value, expected float64, ok bool) {
	t.Helper()
	val, okRes := v.TryF64()
	assert.Equal(t, ok, okRes)
	if ok {
		assert.Equal(t, expected, val)
	}
}

func assertTryDate(t *testing.T, v Value, expected time.Time, ok bool) {
	t.Helper()
	val, okRes := v.TryDate()
	assert.Equal(t, ok, okRes)
	if ok {
		assert.Equal(t, expected, val)
	}
}

func assertTryTimestamp(t *testing.T, v Value, expected int64, ok bool) {
	t.Helper()
	val, okRes := v.TryTimestamp()
	assert.Equal(t, ok, okRes)
	if ok {
		assert.Equal(t, expected, val)
	}
}

func TestValueTypedNull(t *testing.T) {
	v := ValTypedNull(TypeI64)
	assert.Equal(t, TypeI64, v.V)
}

func TestValueExtensions(t *testing.T) {
	v1 := ValI64(123)
	id, _ := v1.EntityIdValue()
	assert.Equal(t, "", id) // int64 stub returns empty string for now
	assert.False(t, v1.TeaqlIsEmpty())
	
	v2 := ValText("hello")
	id2, _ := v2.EntityIdValue()
	assert.Equal(t, "hello", id2)
	assert.False(t, v2.TeaqlIsEmpty())
	
	v3 := ValText("")
	assert.True(t, v3.TeaqlIsEmpty())
	
	v4 := ValNull()
	assert.True(t, v4.TeaqlIsEmpty())
	
	v5 := Value{V: map[string]any{"a": 1}}
	obj, ok := v5.Object()
	assert.True(t, ok)
	assert.Equal(t, 1, obj["a"])
	assert.False(t, v5.TeaqlIsEmpty())
	
	v6 := Value{V: map[string]any{}}
	assert.True(t, v6.TeaqlIsEmpty())
}

func TestValue_TryText_RejectsNonString(t *testing.T) {
	v := ValI64(42)
	s, ok := v.TryText()
	assert.False(t, ok)
	assert.Equal(t, "", s)
}

func TestValue_TryBool(t *testing.T) {
	vTrue := ValBool(true)
	b, ok := vTrue.TryBool()
	assert.True(t, ok)
	assert.True(t, b)

	vFalse := ValBool(false)
	b, ok = vFalse.TryBool()
	assert.True(t, ok)
	assert.False(t, b)

	vInvalid := ValI64(1)
	b, ok = vInvalid.TryBool()
	assert.False(t, ok)
	assert.False(t, b)
}

func TestValue_MoreCoverage(t *testing.T) {
	// EntityIdValue other types
	vInt := ValI64(10)
	id, ok := vInt.EntityIdValue()
	assert.False(t, ok)
	assert.Equal(t, "", id)

	vUint := ValU64(10)
	id, ok = vUint.EntityIdValue()
	assert.False(t, ok)
	assert.Equal(t, "", id)

	vBool := ValBool(true)
	id, ok = vBool.EntityIdValue()
	assert.False(t, ok)
	assert.Equal(t, "", id)

	// Object other type
	vStr := ValText("not object")
	obj, ok := vStr.Object()
	assert.False(t, ok)
	assert.Nil(t, obj)

	// TeaqlIsEmpty slices
	vSliceEmpty := Value{V: []any{}}
	assert.True(t, vSliceEmpty.TeaqlIsEmpty())

	vSliceFull := Value{V: []any{1}}
	assert.False(t, vSliceFull.TeaqlIsEmpty())

	vOther := ValI64(0)
	assert.False(t, vOther.TeaqlIsEmpty())

	// ToJsonValue
	assert.Equal(t, "test", ValText("test").ToJsonValue())
}
