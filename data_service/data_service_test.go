package data_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

func TestMutationRequestTraceAndCommentAccessors(t *testing.T) {
	id1 := uint64(1)
	trace1 := &core.TraceNode{
		EntityType: "User",
		EntityId:   &id1,
		Comment:    "Create User",
	}
	trace2 := &core.TraceNode{
		EntityType: "Profile",
		Comment:    "Create Profile",
	}
	traceChain := []*core.TraceNode{trace1, trace2}

	// Test Insert
	insertCmd := core.NewInsertCommand("User")
	insertCmd.TraceChain = traceChain
	reqInsert := &InsertMutation{Cmd: insertCmd}
	assert.Equal(t, 2, len(reqInsert.TraceChain()))
	assert.Equal(t, trace2, reqInsert.TraceChain()[1])
	assert.Equal(t, "Create Profile", *reqInsert.Comment())

	// Test Update
	updateCmd := core.NewUpdateCommand("User", core.ValI64(1))
	updateCmd.TraceChain = traceChain
	reqUpdate := &UpdateMutation{Cmd: updateCmd}
	assert.Equal(t, 2, len(reqUpdate.TraceChain()))
	assert.Equal(t, "Create Profile", *reqUpdate.Comment())

	// Test Delete
	deleteCmd := core.NewDeleteCommand("User", core.ValI64(1))
	deleteCmd.TraceChain = traceChain
	deleteCmd.SoftDelete = true
	reqDelete := &DeleteMutation{Cmd: deleteCmd}
	assert.Equal(t, 2, len(reqDelete.TraceChain()))
	assert.Equal(t, "Create Profile", *reqDelete.Comment())

	// Test Recover
	recoverCmd := core.NewRecoverCommand("User", core.ValI64(1), 1)
	recoverCmd.TraceChain = traceChain
	reqRecover := &RecoverMutation{Cmd: recoverCmd}
	assert.Equal(t, 2, len(reqRecover.TraceChain()))
	assert.Equal(t, "Create Profile", *reqRecover.Comment())

	// Test Batch
	reqBatch := &BatchMutation{Mutations: []MutationRequest{reqInsert, reqUpdate}}
	assert.Equal(t, 0, len(reqBatch.TraceChain()))
	assert.Nil(t, reqBatch.Comment())

	// Test empty trace chain
	insertEmpty := core.NewInsertCommand("User")
	reqEmpty := &InsertMutation{Cmd: insertEmpty}
	assert.Equal(t, 0, len(reqEmpty.TraceChain()))
	assert.Nil(t, reqEmpty.Comment())
}
