package runtime

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

func TestGraphNodeOperations(t *testing.T) {
	node := &GraphNode{Entity: "User", Values: make(core.Record)}
	
	node.Values["id"] = core.Value{V: int64(123)}
	assert.Equal(t, int64(123), node.Id())
	
	child := node.Child("posts")
	child.Values["id"] = core.Value{V: int64(1)}
	
	child2 := node.Reference("comments", int64(456))
	assert.Equal(t, int64(456), child2.Id())
	
	rels := node.Relations()
	assert.NotNil(t, rels["posts"])
	assert.NotNil(t, rels["comments"])
	
	node.Remove("comments")
	assert.Nil(t, node.Relations()["comments"])
	
	assert.Equal(t, core.EntityGraphOpSave, node.Operation())
	node.IsDeleted = true
	assert.Equal(t, core.EntityGraphOpDelete, node.Operation())
	
	node.Comment("this is a test")
	assert.Equal(t, "this is a test", node.CommentText)
	node.SetComment("updated test")
	assert.Equal(t, "updated test", node.CommentText)
}
