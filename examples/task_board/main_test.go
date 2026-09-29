package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskBoardDefaultLogsHaveIntentAndMaskBusinessValues(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "TEAQL_TASK_BOARD_DB=") && !strings.HasPrefix(entry, "TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "TEAQL_TASK_BOARD_DB="+filepath.Join(t.TempDir(), "task-board.db"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("example failed: %v\n%s", err, output)
	}
	queries, writes := 0, 0
	for _, line := range strings.Split(string(output), "\n") {
		if !strings.HasPrefix(line, "[TeaQL SQL]") {
			continue
		}
		if strings.Contains(line, "[query]") || strings.Contains(line, "[select]") {
			queries++
			if strings.Contains(line, `comment=""`) || strings.Contains(line, `purpose=""`) {
				t.Fatalf("query has empty intent: %s", line)
			}
		} else if strings.Contains(line, "[insert]") || strings.Contains(line, "[update]") || strings.Contains(line, "[delete]") {
			writes++
			if strings.Contains(line, `auditReason=""`) {
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
