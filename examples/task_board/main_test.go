package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskBoardDefaultLogsHaveIntentAndMaskBusinessValues(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "TEAQL_TASK_BOARD_DB=") && !strings.HasPrefix(entry, "TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS=") &&
			!strings.HasPrefix(entry, "TEAQL_LOG_ENDPOINT=") && !strings.HasPrefix(entry, "TEAQL_LOG_FORMAT=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "TEAQL_TASK_BOARD_DB="+filepath.Join(t.TempDir(), "task-board.db"), "TEAQL_LOG_ENDPOINT=stdout")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("example failed: %v\n%s", err, output)
	}
	queries, writes := 0, 0
	for _, line := range strings.Split(string(output), "\n") {
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var record struct {
			Category    string `json:"category"`
			SQL         string `json:"sql"`
			Comment     string `json:"comment"`
			Purpose     string `json:"purpose"`
			AuditReason string `json:"audit_reason"`
		}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("invalid JSON log: %v: %s", err, line)
		}
		if record.Category != "SQL" {
			continue
		}
		if strings.Contains(record.SQL, "SELECT ") {
			queries++
			if strings.TrimSpace(record.Comment) == "" || strings.TrimSpace(record.Purpose) == "" {
				t.Fatalf("query has empty intent: %s", line)
			}
		} else if strings.Contains(record.SQL, "INSERT ") || strings.Contains(record.SQL, "UPDATE ") || strings.Contains(record.SQL, "DELETE ") {
			writes++
			if strings.TrimSpace(record.AuditReason) == "" {
				t.Fatalf("write has empty audit reason: %s", line)
			}
		}
	}
	if queries == 0 || writes == 0 {
		t.Fatalf("missing query/write evidence: %s", output)
	}
	if !strings.Contains(string(output), "PASS task board: governed Q/E/mutation and masked detail") {
		t.Fatalf("missing full flow evidence: %s", output)
	}
	if strings.Contains(string(output), "PRIVATE-TASK-DETAIL") {
		t.Fatal("masked detail reached output")
	}
	if !strings.Contains(string(output), "'PR***************IL' /* masked */") ||
		!strings.Contains(string(output), "'RENAME'") {
		t.Fatal("expected field-level masked detail beside visible ordinary action")
	}
}
