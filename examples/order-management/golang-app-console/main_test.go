package main

import (
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
				!strings.HasPrefix(entry, "TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS=") {
				cmd.Env = append(cmd.Env, entry)
			}
		}
		cmd.Env = append(cmd.Env, "TEAQL_ORDER_MANAGEMENT_DB="+database)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("example run %d failed: %v\n%s", run, err, output)
		}
		queries, writes := 0, 0
		for _, line := range strings.Split(string(output), "\n") {
			if !strings.HasPrefix(line, "[TeaQL SQL]") {
				continue
			}
			if strings.Contains(line, "[query]") || strings.Contains(line, "[select]") {
				queries++
				if strings.Contains(line, `comment=""`) || strings.Contains(line, `purpose=""`) {
					t.Fatalf("run %d query lost intent: %s", run, line)
				}
			} else if strings.Contains(line, "[insert]") || strings.Contains(line, "[update]") || strings.Contains(line, "[delete]") {
				writes++
				if strings.Contains(line, `auditReason=""`) {
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
			customerSQL := ""
			for _, line := range strings.Split(string(output), "\n") {
				if strings.Contains(line, "INSERT INTO customer_data") {
					customerSQL = line
					break
				}
			}
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
