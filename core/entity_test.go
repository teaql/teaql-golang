package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
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
