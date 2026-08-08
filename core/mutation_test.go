package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestMutationCommands(t *testing.T) {
	insert := NewInsertCommand("Order").Value("name", ValText("test"))
	assert.Equal(t, "Order", insert.Entity)
	assert.Equal(t, "test", insert.Values["name"].V)
	
	update := NewUpdateCommand("Order", ValI64(1)).WithExpectedVersion(2).Value("name", ValText("test2"))
	assert.Equal(t, "Order", update.Entity)
	assert.Equal(t, int64(1), update.Id.V)
	assert.Equal(t, int64(2), *update.ExpectedVersion)
	assert.Equal(t, "test2", update.Values["name"].V)
	
	deleteCmd := NewDeleteCommand("Order", ValI64(1)).WithExpectedVersion(2).HardDelete()
	assert.Equal(t, "Order", deleteCmd.Entity)
	assert.Equal(t, int64(1), deleteCmd.Id.V)
	assert.Equal(t, int64(2), *deleteCmd.ExpectedVersion)
	assert.False(t, deleteCmd.SoftDelete)
}

func TestBatchCommands(t *testing.T) {
	batchInsert := NewBatchInsertCommand("User")
	assert.Equal(t, "User", batchInsert.Entity)
	assert.Empty(t, batchInsert.BatchValues)

	batchUpdate := NewBatchUpdateCommand("User", []string{"name", "age"})
	assert.Equal(t, "User", batchUpdate.Entity)
	assert.Equal(t, []string{"name", "age"}, batchUpdate.UpdateFields)
	assert.Empty(t, batchUpdate.BatchValues)
}

func TestRecoverCommand(t *testing.T) {
	recoverCmd := NewRecoverCommand("User", ValI64(10), 5)
	assert.Equal(t, "User", recoverCmd.Entity)
	assert.Equal(t, int64(10), recoverCmd.Id.V)
	assert.Equal(t, int64(5), recoverCmd.ExpectedVersion)
}

