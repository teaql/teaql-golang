package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrderExampleGeneratedIntentAndIdempotence(t *testing.T) {
	database := filepath.Join(t.TempDir(), "order.db")
	for run := 1; run <= 2; run++ {
		cmd := exec.Command("go", "run", ".")
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "TEAQL_ORDER_MANAGEMENT_DB=") &&
				!strings.HasPrefix(entry, "TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS=") &&
				!strings.HasPrefix(entry, "TEAQL_LOG_ENDPOINT=") && !strings.HasPrefix(entry, "TEAQL_LOG_FORMAT=") {
				cmd.Env = append(cmd.Env, entry)
			}
		}
		cmd.Env = append(cmd.Env, "TEAQL_ORDER_MANAGEMENT_DB="+database, "TEAQL_LOG_ENDPOINT=stdout")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("example run %d failed: %v\n%s", run, err, output)
		}
		queries, writes := 0, 0
		customerSQL := ""
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
			if strings.Contains(record.SQL, "INSERT INTO customer_data") {
				customerSQL = record.SQL
			}
			if strings.Contains(record.SQL, "SELECT ") {
				queries++
				if strings.TrimSpace(record.Comment) == "" || strings.TrimSpace(record.Purpose) == "" {
					t.Fatalf("run %d query lost intent: %s", run, line)
				}
			} else if strings.Contains(record.SQL, "INSERT ") || strings.Contains(record.SQL, "UPDATE ") || strings.Contains(record.SQL, "DELETE ") {
				writes++
				if strings.TrimSpace(record.AuditReason) == "" {
					t.Fatalf("run %d write lost audit reason: %s", run, line)
				}
			}
		}
		if queries == 0 {
			t.Fatalf("run %d emitted no governed queries: %s", run, output)
		}
		if run == 1 && writes == 0 {
			t.Fatalf("first run emitted no audited writes: %s", output)
		}
		if strings.Contains(string(output), "masked-in-quick-start") {
			t.Fatalf("customer email reached diagnostics on run %d", run)
		}
		if run == 1 {
			if !strings.Contains(customerSQL, "/* masked */") || !strings.Contains(customerSQL, "'Acme Retail'") {
				t.Fatalf("customer SQL did not preserve ordinary name beside masked email: %s", customerSQL)
			}
		}
		if !strings.Contains(string(output), "[query] matched 1 order(s)") {
			t.Fatalf("run %d lost quick-start result: %s", run, output)
		}
		if run == 2 && !strings.Contains(string(output), "no duplicate rows added") {
			t.Fatalf("second run was not idempotent: %s", output)
		}
	}
}
