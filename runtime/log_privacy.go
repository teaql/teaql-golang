package runtime

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"

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

func credentialLogName(name string) bool {
	name = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, name)
	for _, word := range []string{"password", "passwd", "passphrase", "privatekey", "secret", "accesstoken", "refreshtoken", "idtoken", "apikey", "authorization", "credential", "sessiontoken", "magiclinktoken"} {
		if strings.Contains(name, word) {
			return true
		}
	}
	return false
}

func logValueStrings(value any) []string {
	switch v := value.(type) {
	case core.Value:
		return logValueStrings(v.V)
	case nil:
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

func projectedSQLMetadata(metadata data_service.ExecutionMetadata, allow bool) data_service.ExecutionMetadata {
	// No parameter-to-field provenance: credential-bearing statements are
	// entirely redacted, even when ordinary plaintext debugging is enabled.
	credentials := credentialLogName(metadata.ParameterizedSQL) || credentialLogName(sqlLogText(metadata.DebugQuery)) || credentialLogName(fmt.Sprint(metadata.Parameters))
	if allow && !credentials {
		return metadata
	}
	secrets := logValueStrings(metadata.Parameters)
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	scrub := func(s string) string {
		for _, v := range secrets {
			s = strings.ReplaceAll(s, v, "[REDACTED]")
		}
		return s
	}
	copyText := func(s *string) *string {
		if s == nil {
			return nil
		}
		value := scrub(*s)
		return &value
	}
	sql := metadata.ParameterizedSQL
	if strings.ContainsAny(sql, "'\"`$") || strings.Contains(sql, "--") || strings.Contains(sql, "/*") || strings.IndexFunc(sql, unicode.IsDigit) >= 0 {
		sql = "[REDACTED SQL; NOT REPLAYABLE]"
	}
	metadata.ParameterizedSQL = scrub(sql)
	metadata.Comment, metadata.Purpose, metadata.AuditReason = copyText(metadata.Comment), copyText(metadata.Purpose), copyText(metadata.AuditReason)
	trace := make([]*core.TraceNode, len(metadata.TraceChain))
	for i, node := range metadata.TraceChain {
		if node != nil {
			cloned := *node
			cloned.Comment = scrub(node.Comment)
			trace[i] = &cloned
		}
	}
	metadata.TraceChain = trace
	if len(metadata.Parameters) > 0 {
		metadata.ParameterCount = len(metadata.Parameters)
	}
	metadata.Parameters, metadata.DebugQuery = nil, nil
	return metadata
}
