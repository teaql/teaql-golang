package runtime

import (
	stdcontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func loggingName(value string) *string { return &value }

func newTestLogger(t *testing.T, directory string, environment map[string]string) *RuntimeLogger {
	t.Helper()
	logger, err := NewRuntimeLogger(RuntimeLoggingOptions{WorkingDirectory: directory, Environment: environment})
	if err != nil {
		t.Fatal(err)
	}
	return logger
}

func readLoggingRecords(t *testing.T, paths ...string) []map[string]interface{} {
	t.Helper()
	var records []map[string]interface{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 || data[len(data)-1] != '\n' {
			t.Fatalf("incomplete JSON line in %s", path)
		}
		for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"timestamp", "level", "log_name", "runtime", "event", "message"} {
				if record[field] == nil || record[field] == "" {
					t.Fatalf("missing %s: %s", field, line)
				}
			}
			if _, err := time.Parse(time.RFC3339Nano, record["timestamp"].(string)); err != nil || !strings.HasSuffix(record["timestamp"].(string), "Z") || record["runtime"] != "go" {
				t.Fatalf("invalid UTC/runtime: %s", line)
			}
			records = append(records, record)
		}
	}
	return records
}

func TestRuntimeLoggingDefaultStartupAndTwoStarts(t *testing.T) {
	directory := t.TempDir()
	first := newTestLogger(t, directory, map[string]string{})
	if first.Path != filepath.Join(directory, "logs", "teaql.log") || first.MaxSize != 50*1024*1024 || first.MaxFiles != 7 {
		t.Fatal("incorrect defaults", first)
	}
	if readLoggingRecords(t, first.Path)[0]["event"] != "runtime.start" {
		t.Fatal("file must exist before a query")
	}
	second := newTestLogger(t, directory, map[string]string{})
	second.Write("INFO", "probe.success", "Completed", nil)
	records := readLoggingRecords(t, first.Path)
	if len(records) != 3 || records[1]["event"] != "runtime.start" || records[2]["level"] != "INFO" {
		t.Fatal(records)
	}
}

func TestRuntimeLoggingNamePrecedenceAndOverride(t *testing.T) {
	for _, test := range []struct {
		explicit *string
		domain   string
		models   []string
		expected string
	}{
		{loggingName("configured"), "domain", []string{"model"}, "configured"},
		{nil, "domain", []string{"model"}, "domain"},
		{nil, "", []string{"model"}, "model"},
		{nil, "", []string{"first", "second"}, "teaql"},
		{loggingName("订单"), "", nil, "订单"},
	} {
		directory := t.TempDir()
		logger, err := NewRuntimeLogger(RuntimeLoggingOptions{LogName: test.explicit, ModelNames: test.models,
			WorkingDirectory: directory, Environment: map[string]string{"TEAQL_DOMAIN": test.domain, "TEAQL_LOG_ENDPOINT": "different/output.jsonl"}})
		if err != nil || logger.LogName != test.expected || logger.Path != filepath.Join(directory, "different", "output.jsonl") {
			t.Fatal(logger, err)
		}
		if readLoggingRecords(t, logger.Path)[0]["log_name"] != test.expected {
			t.Fatal("override lost log name")
		}
	}
}

func TestRuntimeLoggingRejectsInvalidConfiguration(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../secret", "a/b", `a\b`, "bad\nname", "NUL", "COM1.txt", "trailing.", "trailing "} {
		if _, err := NewRuntimeLogger(RuntimeLoggingOptions{LogName: &name, Environment: map[string]string{}, WorkingDirectory: t.TempDir()}); err == nil || strings.Contains(err.Error(), name) && name != "" {
			t.Fatalf("must reject without echoing configuration: %q", name)
		}
	}
	for _, value := range []string{"0", "-1", "1.2M", "1TB", "9999999999999999999999G"} {
		if _, err := parseLogSize(value); err == nil {
			t.Fatal(value)
		}
	}
	for value, expected := range map[string]int64{"1": 1, "2K": 2048, "3KB": 3072, "2m": 2097152, "1MB": 1048576, "1G": 1073741824, "1GB": 1073741824} {
		if actual, err := parseLogSize(value); err != nil || actual != expected {
			t.Fatal(value, actual, err)
		}
	}
	for _, value := range []string{"-1", "bad", "1.5", "2147483648"} {
		if _, err := NewRuntimeLogger(RuntimeLoggingOptions{Environment: map[string]string{"TEAQL_LOG_MAX_FILES": value}, WorkingDirectory: t.TempDir()}); err == nil {
			t.Fatal(value)
		}
	}
}

func TestRuntimeLoggingSizeOversizeOrderAndRetention(t *testing.T) {
	logger := newTestLogger(t, t.TempDir(), map[string]string{"TEAQL_LOG_MAX_SIZE": "256", "TEAQL_LOG_MAX_FILES": "2"})
	for i := 0; i < 5; i++ {
		if !logger.Write("INFO", "probe", strings.Repeat("界", 200), map[string]interface{}{"sequence": i}) {
			t.Fatal("write failed")
		}
	}
	for suffix, expected := range map[string]int{"": 4, ".1": 3, ".2": 2} {
		records := readLoggingRecords(t, logger.Path+suffix)
		if len(records) != 1 || records[0]["sequence"] != float64(expected) || records[0]["message"] != strings.Repeat("界", 200) {
			t.Fatal("incorrect archive order or split record", suffix, records)
		}
	}
	if _, err := os.Stat(logger.Path + ".3"); !os.IsNotExist(err) {
		t.Fatal("retained too many archives", err)
	}
	logger = newTestLogger(t, filepath.Dir(filepath.Dir(logger.Path)), map[string]string{"TEAQL_LOG_MAX_SIZE": "256", "TEAQL_LOG_MAX_FILES": "0"})
	logger.Write("INFO", "probe", strings.Repeat("x", 400), nil)
	if archives, err := logger.archives(); err != nil || len(archives) != 0 {
		t.Fatal(archives, err)
	}
	if len(readLoggingRecords(t, logger.Path)) != 1 {
		t.Fatal("zero retention must keep active only")
	}
}

func TestRuntimeLoggingUTCDayAndRestart(t *testing.T) {
	for _, restart := range []bool{false, true} {
		directory := t.TempDir()
		logger := newTestLogger(t, directory, map[string]string{})
		before, _ := os.ReadFile(logger.Path)
		yesterday := time.Now().UTC().Add(-24 * time.Hour)
		if err := os.Chtimes(logger.Path, yesterday, yesterday); err != nil {
			t.Fatal(err)
		}
		if restart {
			logger = newTestLogger(t, directory, map[string]string{})
		} else {
			logger.Write("INFO", "next.day", "Next day", nil)
		}
		archive, _ := os.ReadFile(logger.Path + ".1")
		if string(archive) != string(before) || len(readLoggingRecords(t, logger.Path)) != 1 {
			t.Fatal("UTC rotation lost the preceding active file")
		}
	}
}

func TestRuntimeLoggingOrdinaryUTF8WholeLineByteLimit(t *testing.T) {
	logger := newTestLogger(t, t.TempDir(), map[string]string{"TEAQL_LOG_MAX_SIZE": "1K", "TEAQL_LOG_MAX_FILES": "1000"})
	for i := 0; i < 40; i++ {
		if !logger.Write("INFO", "regular", strings.Repeat("界", 30), map[string]interface{}{"index": i}) {
			t.Fatal("write failed")
		}
	}
	files, err := filepath.Glob(logger.Path + "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 2 {
		t.Fatal("No size rotation")
	}
	indices := map[float64]bool{}
	count := 0
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() <= 0 || info.Size() > 1024 {
			t.Fatalf("File size %d", info.Size())
		}
		for _, record := range readLoggingRecords(t, file) {
			count++
			if index, ok := record["index"].(float64); ok {
				if indices[index] {
					t.Fatal("duplicate index")
				}
				indices[index] = true
			}
		}
	}
	if count != 41 || len(indices) != 40 {
		t.Fatalf("Lost records: %d, %d", count, len(indices))
	}
}

func TestRuntimeLoggingRotationFailurePreservesForeignArchiveAndBusinessOutcome(t *testing.T) {
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	before, err := os.ReadFile(logger.Path)
	if err != nil {
		t.Fatal(err)
	}
	archive := logger.Path + ".1"
	if err := os.Mkdir(archive, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(archive, "marker"), []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	yesterday := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(logger.Path, yesterday, yesterday); err != nil {
		t.Fatal(err)
	}
	context := NewUserContextWithLogger(logger)
	_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
	result := 42
	scope.Success(nil)
	if result != 42 || logger.Write("INFO", "probe", "Completed", nil) {
		t.Fatal("rotation I/O policy")
	}
	after, err := os.ReadFile(logger.Path)
	if err != nil || string(after) != string(before) {
		t.Fatal("Changed active log")
	}
	marker, err := os.ReadFile(filepath.Join(archive, "marker"))
	if err != nil || string(marker) != "preserve" {
		t.Fatal("Changed foreign archive directory")
	}
}

func TestRuntimeLoggingDefaultAuditFileKeepsTypedIdentityAndOriginalBusinessValues(t *testing.T) {
	t.Setenv(plaintextLogEnv, "")
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	context := NewUserContextWithLogger(logger)
	secret := "UNIQUE-LOG-PASSWORD-CANARY"
	event := Created("Customer", core.Record{"id": core.ValI64(42), "password": core.ValText(secret)})
	event.TraceChain = []*core.TraceNode{core.NewTraceNode("Customer", nil, "change target 42 "+secret)}
	if err := context.SendEvent(event); err != nil {
		t.Fatal(err)
	}
	var audit map[string]interface{}
	for _, row := range readLoggingRecords(t, logger.Path) {
		if row["event"] == "audit.event" {
			audit = row
		}
	}
	if audit == nil || audit["level"] != "INFO" {
		t.Fatal("No default audit record")
	}
	want, _ := json.Marshal(core.ValI64(42))
	got, _ := json.Marshal(audit["entity_id"])
	if string(got) != string(want) {
		t.Fatalf("Changed typed identity: %s", got)
	}
	text, err := os.ReadFile(logger.Path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(text), secret) {
		t.Fatal("Credential reached file")
	}
	if event.Values["password"].V != secret || event.Values["id"].V != int64(42) || !strings.Contains(event.TraceChain[0].Comment, secret) {
		t.Fatal("Changed business event")
	}
}

type loggingCorrelationTelemetry struct{ NoopRuntimeTelemetry }
type loggingCorrelationScope struct{}

func (loggingCorrelationTelemetry) Start(ctx stdcontext.Context, _ RuntimeOperation) (stdcontext.Context, RuntimeTelemetryScope) {
	return ctx, loggingCorrelationScope{}
}
func (loggingCorrelationScope) Success(map[string]RuntimeAttributeValue) {}
func (loggingCorrelationScope) Failure(string)                           {}
func (loggingCorrelationScope) Correlation() *data_service.SQLExecutionCorrelation {
	return &data_service.SQLExecutionCorrelation{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef"}
}
func TestRuntimeLoggingOperationKeepsAvailableTraceCorrelation(t *testing.T) {
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	context := NewUserContextWithLogger(logger).WithRuntimeTelemetry(loggingCorrelationTelemetry{})
	_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
	scope.Success(nil)
	context.RecordExecutionMetadata(data_service.ExecutionMetadata{Operation: data_service.OpQuery, SQLCorrelation: &data_service.SQLExecutionCorrelation{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef"}})
	records := readLoggingRecords(t, logger.Path)
	if len(records) != 3 || records[1]["trace_id"] != "0123456789abcdef0123456789abcdef" || records[2]["trace_id"] != "0123456789abcdef0123456789abcdef" {
		t.Fatal("Trace correlation lost")
	}
}

type loggingNativeErrorTelemetry struct {
	NoopRuntimeTelemetry
	kinds *[]string
}
type loggingNativeErrorScope struct{ kinds *[]string }

func (t loggingNativeErrorTelemetry) Start(ctx stdcontext.Context, _ RuntimeOperation) (stdcontext.Context, RuntimeTelemetryScope) {
	return ctx, loggingNativeErrorScope{t.kinds}
}
func (loggingNativeErrorScope) Success(map[string]RuntimeAttributeValue) {}
func (s loggingNativeErrorScope) Failure(kind string)                    { *s.kinds = append(*s.kinds, kind) }
func TestRuntimeLoggingNativeRejectionsKeepOriginalSDKErrorTypes(t *testing.T) {
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	kinds := []string{}
	adapter := &loggingNativeErrorTelemetry{kinds: &kinds}
	context := NewUserContextWithLogger(logger).WithRuntimeTelemetry(adapter)
	errorsToReport := []error{&RuntimeError{Type: "Check", Message: "private-check-value"}, &RuntimeError{Type: "MutationPolicy", Message: "private-policy-value"}, &core.RequestIntentError{Code: "QUERY_PURPOSE_REQUIRED"}, errors.New("private-final-value")}
	for _, original := range errorsToReport {
		_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("mutation", "probe", nil))
		FailRuntimeOperation(scope, original)
	}
	records := readLoggingRecords(t, logger.Path)
	if len(records) != 5 {
		t.Fatal("missing outcome")
	}
	for i, level := range []string{"WARN", "WARN", "WARN", "ERROR"} {
		if records[i+1]["level"] != level {
			t.Fatalf("wrong native severity: %v", records[i+1])
		}
		if kinds[i] != RuntimeErrorType(errorsToReport[i]) {
			t.Fatal("Changed SDK error type")
		}
	}
	if context.RuntimeTelemetry() != adapter {
		t.Fatal("Changed installed adapter")
	}
	text, err := os.ReadFile(logger.Path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(text), "private-") {
		t.Fatal("Raw error value reached log")
	}
}

func TestRuntimeLoggingStartupFailureAndPostStartupBusinessOutcome(t *testing.T) {
	directory := t.TempDir()
	if err := os.MkdirAll(filepath.Join(directory, "logs", "teaql.log"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRuntimeLogger(RuntimeLoggingOptions{WorkingDirectory: directory, Environment: map[string]string{}}); err == nil {
		t.Fatal("startup must fail")
	}
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	context := NewUserContextWithLogger(logger).WithRuntimeTelemetry(NoopRuntimeTelemetry{})
	if err := os.Remove(logger.Path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(logger.Path, 0700); err != nil {
		t.Fatal(err)
	}
	_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
	result := 42
	scope.Success(nil)
	if result != 42 || logger.Write("INFO", "probe", "Completed", nil) {
		t.Fatal("post-startup failure policy")
	}
	original := errors.New("private-value-never-log")
	_, scope = StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
	scope.Failure(RuntimeErrorType(original))
	if original.Error() != "private-value-never-log" {
		t.Fatal("changed business error")
	}
}

func TestRuntimeLoggingOutcomesSQLProjectionAndSPIIdentity(t *testing.T) {
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	telemetry := &NoopRuntimeTelemetry{}
	context := NewUserContextWithLogger(logger).WithRuntimeTelemetry(telemetry)
	if context.RuntimeTelemetry() != telemetry {
		t.Fatal("changed installed extension identity")
	}
	for _, kind := range []string{"", "ValidationError", "ConnectionError"} {
		_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
		if kind == "" {
			scope.Success(nil)
		} else {
			scope.Failure(kind)
		}
		scope.Success(nil) // Exactly one completion record.
	}
	secret := "unprojected-password-raw-sql"
	metadata := data_service.ExecutionMetadata{Operation: data_service.OpQuery, ExecutionOutcome: "retrying", DebugQuery: &secret}
	context.RecordExecutionMetadata(metadata)
	records := readLoggingRecords(t, logger.Path)
	if len(records) != 5 || records[1]["level"] != "INFO" || records[2]["level"] != "WARN" || records[3]["level"] != "ERROR" || records[4]["level"] != "WARN" || records[4]["category"] != "SQL" {
		t.Fatal(records)
	}
	data, _ := os.ReadFile(logger.Path)
	if strings.Contains(string(data), secret) {
		t.Fatal("unprojected SQL reached file")
	}
}

func TestRuntimeLoggingConcurrentWriters(t *testing.T) {
	directory := t.TempDir()
	environment := map[string]string{"TEAQL_LOG_MAX_SIZE": "1K", "TEAQL_LOG_MAX_FILES": "1000"}
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		logger := newTestLogger(t, directory, environment)
		workers.Add(1)
		go func(id int) {
			defer workers.Done()
			for i := 0; i < 30; i++ {
				if !logger.Write("INFO", "concurrent", "Completed", map[string]interface{}{"id": fmt.Sprintf("%d/%d", id, i)}) {
					t.Error("concurrent write failed")
				}
			}
		}(worker)
	}
	workers.Wait()
	assertLoggingIDs(t, filepath.Join(directory, "logs", "teaql.log"), 240, 8)
}

func assertLoggingIDs(t *testing.T, path string, expected, starts int) {
	t.Helper()
	paths, err := filepath.Glob(path + "*")
	if err != nil {
		t.Fatal(err)
	}
	records := readLoggingRecords(t, paths...)
	ids := map[string]bool{}
	for _, record := range records {
		if id, ok := record["id"].(string); ok {
			if ids[id] {
				t.Fatal("duplicate event", id)
			}
			ids[id] = true
		}
	}
	if len(ids) != expected || len(records) != expected+starts {
		t.Fatal(len(ids), len(records))
	}
}

func TestRuntimeLoggingProcessHelper(t *testing.T) {
	if os.Getenv("TEAQL_LOGGING_TEST_CHILD") != "yes" {
		return
	}
	logger := newTestLogger(t, os.Getenv("TEAQL_LOGGING_TEST_DIRECTORY"), map[string]string{"TEAQL_LOG_MAX_SIZE": "1K", "TEAQL_LOG_MAX_FILES": "1000"})
	for i := 0; i < 30; i++ {
		if !logger.Write("INFO", "concurrent", "Completed", map[string]interface{}{"id": fmt.Sprintf("%s/%d", os.Getenv("TEAQL_LOGGING_TEST_WORKER"), i)}) {
			t.Fatal("child write")
		}
	}
}

func TestRuntimeLoggingSeparateProcesses(t *testing.T) {
	directory := t.TempDir()
	var commands []*exec.Cmd
	for worker := 0; worker < 4; worker++ {
		command := exec.Command(os.Args[0], "-test.run=^TestRuntimeLoggingProcessHelper$", "-test.timeout=30s")
		command.Env = []string{"TEAQL_LOGGING_TEST_CHILD=yes", "TEAQL_LOGGING_TEST_DIRECTORY=" + directory, fmt.Sprintf("TEAQL_LOGGING_TEST_WORKER=%d", worker), "TEAQL_TELEMETRY_ENABLED=false"}
		command.Stdout, command.Stderr = os.Stdout, os.Stderr
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		commands = append(commands, command)
	}
	t.Cleanup(func() {
		for _, command := range commands {
			if command.ProcessState == nil {
				_ = command.Process.Kill()
				_ = command.Wait()
			}
		}
	})
	for _, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	assertLoggingIDs(t, filepath.Join(directory, "logs", "teaql.log"), 120, 4)
}
