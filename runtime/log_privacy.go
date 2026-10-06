package runtime

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/teaql/teaql-golang/internal/logprivacy"
	"github.com/teaql/teaql-golang/sql"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

const plaintextLogEnv = "TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS"
const plaintextLogAck = "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK"

var plaintextLogWarning sync.Once

func plaintextLogsEnabled() bool {
	enabled := os.Getenv(plaintextLogEnv) == plaintextLogAck
	if enabled {
		plaintextLogWarning.Do(func() {
			log.Print("TeaQL: sensitive plaintext logging enabled; application data may be written to disk. Authentication secrets remain redacted.")
		})
	}
	return enabled
}

func credentialLogName(name string) bool { return logprivacy.CredentialName(name) }

func logValueStrings(value any) []string {
	switch v := value.(type) {
	case core.Value:
		return logValueStrings(v.V)
	case nil:
		return nil
	case core.DataType: // typed SQL NULL carries a type tag, not a business value
		return nil
	case map[string]any:
		var result []string
		for _, child := range v {
			result = append(result, logValueStrings(child)...)
		}
		return result
	case core.Record:
		var result []string
		for _, child := range v {
			result = append(result, logValueStrings(child)...)
		}
		return result
	case []core.Value:
		var result []string
		for _, child := range v {
			result = append(result, logValueStrings(child)...)
		}
		return result
	case []any:
		var result []string
		for _, child := range v {
			result = append(result, logValueStrings(child)...)
		}
		return result
	default:
		text := fmt.Sprint(v)
		if text == "" {
			return nil
		}
		return []string{text}
	}
}

func logHasCredentials(value any) bool {
	switch v := value.(type) {
	case core.Value:
		return logHasCredentials(v.V)
	case core.Record:
		for key, child := range v {
			if credentialLogName(key) || logHasCredentials(child) {
				return true
			}
		}
	case map[string]any:
		for key, child := range v {
			if credentialLogName(key) || logHasCredentials(child) {
				return true
			}
		}
	case []core.Value:
		for _, child := range v {
			if logHasCredentials(child) {
				return true
			}
		}
	case []any:
		for _, child := range v {
			if logHasCredentials(child) {
				return true
			}
		}
	}
	return false
}

func cloneLogValue(value core.Value) core.Value {
	return core.CloneValue(value)
}

func businessMaskValue(value core.Value) core.Value {
	if value.V == nil {
		return core.ValNull()
	}
	if _, typedNull := value.V.(core.DataType); typedNull {
		return value
	}
	switch v := value.V.(type) {
	case []core.Value:
		values := make([]core.Value, len(v))
		for i, child := range v {
			values[i] = businessMaskValue(child)
		}
		value.V = values
		return value
	case core.Record, map[string]any, []any:
		return core.ValText("[REDACTED]")
	}
	return core.ValText(MaskAuditValue(fmt.Sprint(value.V)))
}

var numberedLogBind = regexp.MustCompile(`\$[0-9]+`)
var unsafeLogLiteral = regexp.MustCompile("['\"`$]|--|/\\*|\\b[0-9]+\\b|:[A-Za-z_]")

func bindingLogPolicy(metadata data_service.ExecutionMetadata, index int) string {
	if len(metadata.MaskedParameters) != 0 && len(metadata.MaskedParameters) != len(metadata.Parameters) {
		return "unknown"
	}
	valid := metadata.ParameterLogPolicies != nil && len(metadata.ParameterLogPolicies) == len(metadata.Parameters)
	policy := "unknown"
	if valid {
		policy = metadata.ParameterLogPolicies[index]
	}
	if credentialLogName(metadata.ParameterizedSQL) && (!metadata.GeneratedSQL || !valid) || logHasCredentials(metadata.Parameters[index]) {
		return "credential"
	}
	switch policy {
	case "plain", "masked", "credential":
		return policy
	default:
		return "unknown"
	}
}

func bindingIsMasked(policy string, allow bool) bool {
	return policy == "credential" || policy == "unknown" || (!allow && policy != "plain")
}

// Readback adds its write source to an existing batch source. Resolve every
// layer with the same binding policy; never install this state on UserContext.
func inheritedIntentSecrets(inherited logprivacy.IntentSource, allow bool) []string {
	var secrets []string
	var visit func(logprivacy.IntentSource, bool)
	visit = func(inherited logprivacy.IntentSource, allow bool) {
		var sources []data_service.ExecutionMetadata
		switch source := logprivacy.ReadIntentSource(inherited).(type) {
		case data_service.ExecutionMetadata:
			sources = []data_service.ExecutionMetadata{source}
		case []data_service.ExecutionMetadata:
			sources = source
		}
		for _, source := range sources {
			allowSource := allow && source.LogMode != "masked"
			for index, value := range source.Parameters {
				if bindingIsMasked(bindingLogPolicy(source, index), allowSource) {
					secrets = append(secrets, logValueStrings(value)...)
				}
			}
			visit(source.InheritedIntent, allowSource)
		}
	}
	visit(inherited, allow)
	return secrets
}

func projectedSQLMetadata(metadata data_service.ExecutionMetadata, allow bool) data_service.ExecutionMetadata {
	original := metadata
	// Safe projections never become raw values again on a later opt-in.
	allow = allow && metadata.LogMode != "masked"
	debugSource := metadata.LogMode == "debug-plaintext" || (metadata.DebugQuery != nil && strings.HasPrefix(*metadata.DebugQuery, "-- TeaQL DEBUG PLAINTEXT; EXPLICIT OPT-IN"))
	var remembered data_service.ExecutionMetadata
	var hasRemembered bool
	if debugSource {
		remembered, hasRemembered = rememberedSQLProjection(metadata)
	}
	if !allow && hasRemembered {
		return remembered
	}
	orphanedDebug := debugSource && !allow && !hasRemembered
	flagsValid := len(metadata.MaskedParameters) == 0 || len(metadata.MaskedParameters) == len(metadata.Parameters)
	policiesValid := metadata.ParameterLogPolicies == nil || len(metadata.ParameterLogPolicies) == len(metadata.Parameters)
	// Generated bindings classify credentials individually. A credential column in
	// the SELECT list must not override unrelated ordinary/masked field policies.
	credentials := credentialLogName(metadata.ParameterizedSQL) &&
		(!metadata.GeneratedSQL || metadata.ParameterLogPolicies == nil || !policiesValid)
	policies := make([]string, len(metadata.Parameters))
	masked := make([]bool, len(policies))
	values := make([]core.Value, len(policies))
	var secrets []string
	for i, value := range metadata.Parameters {
		policy := bindingLogPolicy(metadata, i)
		policies[i] = policy
		masked[i] = bindingIsMasked(policy, allow)
		if masked[i] {
			secrets = append(secrets, logValueStrings(value)...)
			_, typedNull := value.V.(core.DataType)
			if value.V == nil || typedNull {
				values[i] = value
			} else if policy == "masked" {
				values[i] = businessMaskValue(value)
			} else {
				values[i] = core.ValText("[REDACTED]")
			}
		} else {
			values[i] = cloneLogValue(value)
		}
	}
	// Readback binds only an ID, but its inherited intent can mention sensitive
	// write values. Reuse precisely the same policy resolution for those values.
	secrets = append(secrets, inheritedIntentSecrets(metadata.InheritedIntent, allow)...)
	intentSecrets := append([]string(nil), secrets...)
	if targetID, ok := logprivacy.ReadIntentSource(metadata.IntentTargetID).(core.Value); ok {
		intentSecrets = append(intentSecrets, logValueStrings(targetID)...)
	}
	metadata.InheritedIntent = logprivacy.IntentSource{}
	metadata.IntentTargetID = logprivacy.IntentSource{}
	metadata.LogProjection = logprivacy.ProjectionState{}
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	sort.Slice(intentSecrets, func(i, j int) bool { return len(intentSecrets[i]) > len(intentSecrets[j]) })
	scrubWith := func(s string, values []string) string {
		if orphanedDebug && s != "" {
			return "[REDACTED]"
		}
		for _, v := range values {
			s = strings.ReplaceAll(s, v, "[REDACTED]")
		}
		return s
	}
	scrub := func(s string) string { return scrubWith(s, secrets) }
	scrubIntent := func(s string) string { return scrubWith(s, intentSecrets) }
	copyText := func(s *string) *string {
		if s == nil {
			return nil
		}
		value := scrub(*s)
		return &value
	}
	copyIntent := func(s *string) *string {
		if s == nil {
			return nil
		}
		value := scrubIntent(*s)
		return &value
	}
	unsafe := (!allow || credentials) && !metadata.GeneratedSQL && unsafeLogLiteral.MatchString(numberedLogBind.ReplaceAllString(metadata.ParameterizedSQL, "?"))
	omitted := "[REDACTED SQL; NOT REPLAYABLE]"
	rendered := omitted
	previousOmission := metadata.OmissionReason
	metadata.OmissionReason = ""
	if previousOmission != "" {
		switch previousOmission {
		case "untrusted-literal-sql", "policy-count-mismatch", "mask-count-mismatch", "unsupported-or-mismatched-bindings", "unavailable-sql":
			metadata.OmissionReason = previousOmission
		default:
			metadata.OmissionReason = "unavailable-sql"
		}
	} else if unsafe {
		metadata.OmissionReason = "untrusted-literal-sql"
	} else if !policiesValid {
		metadata.OmissionReason = "policy-count-mismatch"
	} else if !flagsValid {
		metadata.OmissionReason = "mask-count-mismatch"
	} else {
		kind := sql.DatabaseKindSQLite
		if metadata.Backend == "postgresql" || (metadata.Backend == "" && numberedLogBind.MatchString(metadata.ParameterizedSQL)) {
			kind = sql.DatabaseKindPostgreSQL
		}
		if metadata.Backend == "mysql" {
			kind = sql.DatabaseKindMySQL
		}
		value, err := sql.RenderSQLLog(metadata.ParameterizedSQL, values, masked, kind)
		if err != nil {
			metadata.OmissionReason = "unsupported-or-mismatched-bindings"
		} else {
			prefix := "-- TeaQL MASKED; NOT REPLAYABLE\n"
			if allow {
				prefix = "-- TeaQL DEBUG PLAINTEXT; EXPLICIT OPT-IN\n"
				for _, flag := range masked {
					if flag {
						prefix = "-- TeaQL DEBUG PLAINTEXT; EXPLICIT OPT-IN; PARTIALLY MASKED; NOT REPLAYABLE\n"
						break
					}
				}
			}
			rendered = prefix + value
		}
	}
	if unsafe {
		metadata.ParameterizedSQL = omitted
	} else if !metadata.GeneratedSQL {
		// Trusted compiled SQL contains bindings, not their values. Substring
		// scrubbing here could replace the digit 1 in LIMIT 10000 or a table name
		// that happens to equal a sensitive value, corrupting repeat projection.
		metadata.ParameterizedSQL = scrub(metadata.ParameterizedSQL)
	}
	metadata.Comment, metadata.Purpose, metadata.AuditReason = copyIntent(metadata.Comment), copyIntent(metadata.Purpose), copyIntent(metadata.AuditReason)
	metadata.BackendRequestId = copyText(metadata.BackendRequestId)
	projectTrace := func(source []*core.TraceNode) []*core.TraceNode {
		trace := core.CloneTraceNodes(source)
		for _, node := range trace {
			if node != nil {
				node.Comment = scrubIntent(node.Comment)
				node.Name = scrubIntent(node.Name)
				node.EntityType = scrubIntent(node.EntityType)
				node.Kind = scrubIntent(node.Kind)
			}
		}
		return trace
	}
	metadata.TraceChain = projectTrace(metadata.TraceChain)
	metadata.MutationLineage = projectTrace(metadata.MutationLineage)
	metadata.ParameterCount = len(metadata.Parameters)
	metadata.Parameters = values
	metadata.ParameterLogPolicies, metadata.MaskedParameters = policies, masked
	metadata.DebugQuery = &rendered
	metadata.LogMode = "masked"
	if allow {
		metadata.LogMode = "debug-plaintext"
		if !hasRemembered {
			remembered = projectedSQLMetadata(original, false)
		}
		metadata = rememberSQLProjection(metadata, remembered)
	}
	return metadata
}
