package data_service

import (
	"strings"

	"github.com/teaql/teaql-golang/core"
)

// Native batches own intent but have no aggregate entity/ID. Capture the root
// reason on each already-owned command using its first responsibility type.
// Execution, readback and committed audit then consume the same item lineage.
func prefixBatchResponsibility(request MutationRequest, reason string) {
	var lineage *[]*core.TraceNode
	var entity string
	var target core.Value
	switch item := request.(type) {
	case *InsertMutation:
		lineage, entity, target = &item.Cmd.TraceChain, item.Cmd.Entity, item.Cmd.Values["id"]
	case *UpdateMutation:
		lineage, entity, target = &item.Cmd.TraceChain, item.Cmd.Entity, item.Cmd.Id
	case *DeleteMutation:
		lineage, entity, target = &item.Cmd.TraceChain, item.Cmd.Entity, item.Cmd.Id
	case *RecoverMutation:
		lineage, entity, target = &item.Cmd.TraceChain, item.Cmd.Entity, item.Cmd.Id
	default:
		// Nested batches normalized their own children during recursive capture.
		return
	}
	root := core.NewTypedTraceNode("auditReason", entity, reason)
	for index, node := range *lineage {
		if node == nil || !strings.EqualFold(strings.ReplaceAll(node.Kind, "_", ""), "auditreason") {
			continue
		}
		if node.Comment == reason {
			// Capture can be repeated at adapter/transaction boundaries. One
			// semantic root slot keeps its known ID, never a duplicate reason.
			if index == 0 {
				return
			}
			root = node
			*lineage = append((*lineage)[:index], (*lineage)[index+1:]...)
		} else {
			root.EntityType, root.Name = node.EntityType, node.EntityType
			if root.EntityType == "" {
				root.EntityType, root.Name = node.Name, node.Name
			}
		}
		*lineage = append([]*core.TraceNode{root}, *lineage...)
		return
	}
	// With no local reason, the item itself supplies the known entity identity;
	// there is still no synthetic container or cross-item identity.
	if id, ok := target.TryU64(); ok {
		root.EntityId = &id
	}
	*lineage = append([]*core.TraceNode{root}, *lineage...)
}
