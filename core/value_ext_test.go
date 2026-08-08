package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValueExt_List(t *testing.T) {
	list := []Value{ValI64(1), ValI64(2)}
	v := ValList(list)
	
	l, ok := v.TryList()
	assert.True(t, ok)
	assert.Equal(t, list, l)
	
	v2 := ValI64(1)
	l2, ok2 := v2.TryList()
	assert.False(t, ok2)
	assert.Nil(t, l2)
}

func TestValueExt_Object(t *testing.T) {
	obj := Record{"a": ValI64(1)}
	v := ValObject(obj)
	
	o, ok := v.TryObject()
	assert.True(t, ok)
	assert.Equal(t, obj, o)
	
	v2 := ValI64(1)
	o2, ok2 := v2.TryObject()
	assert.False(t, ok2)
	assert.Nil(t, o2)
}

func TestValueExt_Json(t *testing.T) {
	v := ValJson(map[string]any{"a": 1})
	j, ok := v.TryJson()
	assert.True(t, ok)
	assert.NotNil(t, j)
	
	vNull := ValNull()
	j2, ok2 := vNull.TryJson()
	assert.False(t, ok2)
	assert.Nil(t, j2)
}

func TestValueExt_IsNull(t *testing.T) {
	vNull := ValNull()
	assert.True(t, vNull.IsNull())
	
	vNotNull := ValI64(1)
	assert.False(t, vNotNull.IsNull())
}
