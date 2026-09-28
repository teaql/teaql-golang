package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

type sqlProjectionState struct {
	fingerprint [32]byte
	safe        ds.ExecutionMetadata
}

func sqlProjectionFingerprint(metadata ds.ExecutionMetadata) ([32]byte, bool) {
	// Internal source/state fields are excluded by their JSON tags. Hash rather
	// than retaining a second plaintext debug record. Include native scalar text
	// and types so JSON numeric normalization cannot alias a changed binding.
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return [32]byte{}, false
	}
	hash := sha256.New()
	hash.Write(encoded)
	for _, value := range metadata.Parameters {
		fmt.Fprintf(hash, "\x00%T:%v", value.V, value.V)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, true
}

func rememberedSQLProjection(metadata ds.ExecutionMetadata) (ds.ExecutionMetadata, bool) {
	state, ok := logprivacy.ReadProjectionState(metadata.LogProjection).(*sqlProjectionState)
	if !ok || state == nil {
		return ds.ExecutionMetadata{}, false
	}
	fingerprint, valid := sqlProjectionFingerprint(metadata)
	if !valid || fingerprint != state.fingerprint {
		return ds.ExecutionMetadata{}, false
	}
	return cloneSQLProjection(state.safe), true
}

func rememberSQLProjection(metadata, safe ds.ExecutionMetadata) ds.ExecutionMetadata {
	fingerprint, valid := sqlProjectionFingerprint(metadata)
	if !valid {
		metadata.LogProjection = logprivacy.ProjectionState{}
		return metadata
	}
	metadata.LogProjection = logprivacy.NewProjectionState(&sqlProjectionState{fingerprint: fingerprint, safe: cloneSQLProjection(safe)})
	return metadata
}

func cloneSQLProjection(metadata ds.ExecutionMetadata) ds.ExecutionMetadata {
	cloneText := func(value *string) *string {
		if value == nil {
			return nil
		}
		result := *value
		return &result
	}
	metadata.Comment = cloneText(metadata.Comment)
	metadata.Purpose = cloneText(metadata.Purpose)
	metadata.AuditReason = cloneText(metadata.AuditReason)
	metadata.BackendRequestId = cloneText(metadata.BackendRequestId)
	metadata.DebugQuery = cloneText(metadata.DebugQuery)
	if metadata.ResultCount != nil {
		value := *metadata.ResultCount
		metadata.ResultCount = &value
	}
	if metadata.AffectedRows != nil {
		value := *metadata.AffectedRows
		metadata.AffectedRows = &value
	}
	parameters := make([]core.Value, len(metadata.Parameters))
	for i, value := range metadata.Parameters {
		parameters[i] = cloneLogValue(value)
	}
	metadata.Parameters = parameters
	if metadata.ParameterLogPolicies != nil {
		policies := make([]string, len(metadata.ParameterLogPolicies))
		copy(policies, metadata.ParameterLogPolicies)
		metadata.ParameterLogPolicies = policies
	}
	if metadata.MaskedParameters != nil {
		flags := make([]bool, len(metadata.MaskedParameters))
		copy(flags, metadata.MaskedParameters)
		metadata.MaskedParameters = flags
	}
	trace := make([]*core.TraceNode, len(metadata.TraceChain))
	for i, node := range metadata.TraceChain {
		if node != nil {
			value := *node
			trace[i] = &value
		}
	}
	metadata.TraceChain = trace
	metadata.InheritedIntent = logprivacy.IntentSource{}
	metadata.LogProjection = logprivacy.ProjectionState{}
	return metadata
}
