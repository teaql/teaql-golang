package core

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseEntityDataToAndFromRecord(t *testing.T) {
	b := NewBaseEntityData().
		WithId(7).
		WithVersion(2).
		WithDynamic("name", ValText("test"))

	record := b.ToRecord()

	valId, ok := record["id"].TryU64()
	assert.True(t, ok)
	assert.Equal(t, uint64(7), valId)

	valVersion, ok := record["version"].TryI64()
	assert.True(t, ok)
	assert.Equal(t, int64(2), valVersion)

	valName, ok := record["name"].TryText()
	assert.True(t, ok)
	assert.Equal(t, "test", valName)

	b2, err := BaseEntityDataFromRecord(record)
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), b2.Id)
	assert.Equal(t, int64(2), b2.Version)

	name, ok := b2.DynamicText("name")
	assert.True(t, ok)
	assert.Equal(t, "test", name)
}

type entityTestMockEntity struct {
	comment string
}

func (m *entityTestMockEntity) EntityName() string                  { return "Mock" }
func (m *entityTestMockEntity) EntityDescriptor() *EntityDescriptor { return nil }
func (m *entityTestMockEntity) FromRecord(record Record) error      { return nil }
func (m *entityTestMockEntity) IntoRecord() Record                  { return nil }
func (m *entityTestMockEntity) DirtyFields() []string               { return nil }
func (m *entityTestMockEntity) IsMarkedAsDelete() bool              { return false }
func (m *entityTestMockEntity) IsNew() bool                         { return false }
func (m *entityTestMockEntity) MarkAsNew()                          {}
func (m *entityTestMockEntity) GetComment() *string                 { return &m.comment }
func (m *entityTestMockEntity) SetComment(comment string)           { m.comment = comment }
func (m *entityTestMockEntity) OriginalValues() Record              { return nil }
func (m *entityTestMockEntity) OnLoaded(context any)                {}
func (m *entityTestMockEntity) IntoJson() any                       { return nil }

func TestEntityError(t *testing.T) {
	err := NewEntityError("TestEntity", "test message")
	assert.Equal(t, "TestEntity: test message", err.Error())
}

func TestAuditedSuccess(t *testing.T) {
	me := &entityTestMockEntity{}
	audited := NewAudited(me, "test comment")
	assert.Equal(t, "test comment", audited.GetComment())
	assert.Equal(t, me, audited.Entity())
	assert.Equal(t, "test comment", *audited.IntoEntity().GetComment())
	assert.Equal(t, "test comment", me.comment) // IntoEntity sets comment
}

func TestAuditedPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewAudited(&entityTestMockEntity{}, "   ")
	})
}

func TestBaseEntityDataDynamics(t *testing.T) {
	b := NewBaseEntityData()

	// DynamicI64
	b.PutDynamic("i64", ValI64(42))
	vI64, ok := b.DynamicI64("i64")
	assert.True(t, ok)
	assert.Equal(t, int64(42), vI64)

	_, ok = b.DynamicI64("missing")
	assert.False(t, ok)

	// DynamicU64
	b.PutDynamic("u64", ValU64(43))
	vU64, ok := b.DynamicU64("u64")
	assert.True(t, ok)
	assert.Equal(t, uint64(43), vU64)

	_, ok = b.DynamicU64("missing")
	assert.False(t, ok)

	// DynamicDecimal
	dec := decimal.NewFromFloat(44.4)
	b.PutDynamic("dec", ValDecimal(dec))
	vDec, ok := b.DynamicDecimal("dec")
	assert.True(t, ok)
	assert.True(t, dec.Equal(vDec))

	_, ok = b.DynamicDecimal("missing")
	assert.False(t, ok)

	// DynamicF64
	b.PutDynamic("f64", ValF64(45.5))
	vF64, ok := b.DynamicF64("f64")
	assert.True(t, ok)
	assert.Equal(t, float64(45.5), vF64)

	_, ok = b.DynamicF64("missing")
	assert.False(t, ok)

	// DynamicText
	_, ok = b.DynamicText("missing")
	assert.False(t, ok)

	// DynamicBool
	b.PutDynamic("bool", ValBool(true))
	vBool, ok := b.DynamicBool("bool")
	assert.True(t, ok)
	assert.Equal(t, true, vBool)

	_, ok = b.DynamicBool("missing")
	assert.False(t, ok)

	// RemoveDynamic
	b.RemoveDynamic("bool")
	_, ok = b.DynamicBool("bool")
	assert.False(t, ok)
}

func TestBaseEntityDataFromRecordErrors(t *testing.T) {
	// invalid id field
	_, err := BaseEntityDataFromRecord(Record{"id": ValText("not-id")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id field")

	// negative id field
	_, err = BaseEntityDataFromRecord(Record{"id": ValI64(-1)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id field")

	// empty record (missing id and version)
	bEmpty, err := BaseEntityDataFromRecord(Record{})
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), bEmpty.Id)
	assert.Equal(t, int64(0), bEmpty.Version)

	// id is I64
	b, err := BaseEntityDataFromRecord(Record{"id": ValI64(10)})
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), b.Id)

	// invalid version field
	_, err = BaseEntityDataFromRecord(Record{"version": ValText("not-version")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version field")
}
