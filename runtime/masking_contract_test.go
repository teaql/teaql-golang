package runtime

import (
	"bytes"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"os"
	"strings"
	"testing"
)

func TestMaskGolden(t *testing.T) {
	t.Setenv(plaintextLogEnv, "")
	data, err := os.ReadFile("../test-vectors/masking-v1.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[1:] {
		v := strings.Split(line, "\t")
		t.Run(v[0], func(t *testing.T) {
			if got := MaskAuditValue(v[1]); got != v[2] {
				t.Fatalf("got %q want %q", got, v[2])
			}
			field := BuildSafeAuditField("name", &v[1], []string{"name"}, nil)
			if !field.Masked || field.Value == nil || *field.Value != v[2] {
				t.Fatalf("audit mask mismatch: %+v", field)
			}
			event := Created("Customer", core.Record{"name": core.ValText(v[1])})
			safe := event.BuildSafeEvent([]string{"name"}, nil)
			if len(safe.Fields) != 1 || safe.Fields[0].Value == nil || *safe.Fields[0].Value != v[2] {
				t.Fatalf("actual audit event mask mismatch: %+v", safe)
			}
		})
	}
}

var maskContractCases = [][2]string{
	{"", ""}, {"Ada", "***"}, {"12345678", "********"},
	{"ABCDEFGH", "AB****GH"}, {"Riverside", "Ri*****de"}, {"O'Reilly", "O'****ly"},
}

func maskContractEntry(raw string) data_service.ExecutionMetadata {
	sql := "UPDATE customer SET name = '" + strings.ReplaceAll(raw, "'", "''") + "'"
	return data_service.ExecutionMetadata{ParameterizedSQL: "UPDATE customer SET name = ?",
		Parameters: []core.Value{core.ValText(raw)}, DebugQuery: &sql}
}
func TestMaskContractLegacyAlgorithm(t *testing.T) {
	for _, c := range maskContractCases {
		if actual := MaskAuditValue(c[0]); actual != c[1] {
			t.Errorf("%q: got %q want %q", c[0], actual, c[1])
		}
	}
}
func TestMaskContractExpandedSql(t *testing.T) {
	t.Setenv(plaintextLogEnv, "")
	for _, c := range maskContractCases {
		t.Run(c[0], func(t *testing.T) {
			raw := maskContractEntry(c[0])
			var output bytes.Buffer
			NewTextDiagnosticSQLLogSink(&output).WriteSQLLog(raw)
			log := output.String()
			if raw.Parameters[0].V != c[0] {
				t.Error("execution value changed")
			}
			// No field policy is attached; unknown values must be fully hidden.
			expected := "name = '"
			if c[0] != "" && strings.Contains(log, "'"+strings.ReplaceAll(c[0], "'", "''")+"'") {
				t.Error("raw parameter leaked")
			}
			if len(c[0]) >= 8 && !strings.HasPrefix(c[1], "*") && strings.Contains(log, strings.ReplaceAll(c[1], "'", "''")) {
				t.Error("unknown parameter must not expose partial mask")
			}
			if !strings.Contains(log, expected) || !strings.Contains(strings.ToLower(log), "masked") ||
				strings.Contains(log, "name = ?") || strings.Contains(log, "[REDACTED SQL") {
				t.Errorf("must retain expanded masked SQL %q; got %s", expected, log)
			}
		})
	}
}
func TestMaskContractDebugProvenance(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	for i := 0; i < 2; i++ {
		var output bytes.Buffer
		business := maskContractEntry("Riverside")
		business.ParameterLogPolicies = []string{"masked"}
		NewSensitiveDiagnosticSQLLogSink(&output).WriteSQLLog(business)
		log := output.String()
		if !strings.Contains(log, "'Riverside'") || !strings.Contains(strings.ToUpper(log), "PLAINTEXT") ||
			!strings.Contains(strings.ToUpper(log), "DEBUG") {
			t.Errorf("record %d lacks debug provenance: %s", i, log)
		}
	}
}
