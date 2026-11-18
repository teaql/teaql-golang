package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

var presentationHeader = regexp.MustCompile(`^(\S+) \[TeaQL\]\[(.*)\]\[(INFO|WARN|ERROR)\] (.*) runtime=(\S+)$`)

func presentationScalar(t *testing.T, value string) string {
	t.Helper()
	decoded, err := strconv.Unquote(`"` + strings.ReplaceAll(value, `"`, `\"`) + `"`)
	if err != nil {
		t.Fatal("invalid scalar escaping", err)
	}
	return decoded
}

func presentationRecords(t *testing.T, format string, data []byte) []map[string]interface{} {
	t.Helper()
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatal("incomplete event")
	}
	var output []map[string]interface{}
	var sql []string
	for _, line := range strings.Split(string(data[:len(data)-1]), "\n") {
		if format == "json" {
			var row map[string]interface{}
			decoder := json.NewDecoder(strings.NewReader(line))
			decoder.UseNumber()
			if err := decoder.Decode(&row); err != nil {
				t.Fatal(err)
			}
			output = append(output, row)
			continue
		}
		for _, c := range line {
			if unicode.IsControl(c) || c == '\u2028' || c == '\u2029' {
				t.Fatal("unescaped human control")
			}
		}
		if header := presentationHeader.FindStringSubmatch(line); header != nil {
			date, err := time.Parse(time.RFC3339Nano, header[1])
			if err != nil || date.Location() != time.UTC {
				t.Fatal("non-UTC header", header)
			}
			output = append(output, map[string]interface{}{"timestamp": header[1], "log_name": presentationScalar(t, header[2]),
				"level": header[3], "event": presentationScalar(t, header[4]), "runtime": header[5]})
			sql = nil
			continue
		}
		if len(output) == 0 || !strings.HasPrefix(line, "  ") {
			t.Fatal("forged header or interleaved block", line)
		}
		row := output[len(output)-1]
		if _, ok := row["message"]; !ok {
			row["message"] = presentationScalar(t, line[2:])
		} else if line == "  sql:" {
			sql = []string{}
			row["sql"] = ""
		} else if strings.HasPrefix(line, "    ") && sql != nil {
			sql = append(sql, presentationScalar(t, line[4:]))
			row["sql"] = strings.Join(sql, "\n")
		} else {
			parts := strings.SplitN(line[2:], "=", 2)
			if len(parts) != 2 {
				t.Fatal("missing typed detail", line)
			}
			decoder := json.NewDecoder(strings.NewReader(parts[1]))
			decoder.UseNumber()
			var value interface{}
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			row[presentationScalar(t, parts[0])] = value
			sql = nil
		}
	}
	for _, row := range output {
		for _, key := range []string{"timestamp", "level", "log_name", "runtime", "event", "message"} {
			if _, ok := row[key].(string); !ok {
				t.Fatal("missing event field", key, row)
			}
		}
	}
	return output
}

func presentationRead(t *testing.T, format string, paths ...string) []map[string]interface{} {
	t.Helper()
	var rows []map[string]interface{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, presentationRecords(t, format, data)...)
	}
	return rows
}

func TestRuntimeLoggingPresentationsSelectionAndAppend(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		t.Run(format, func(t *testing.T) {
			for index, test := range []struct {
				explicit *string
				domain   string
				models   []string
				name     string
			}{
				{loggingName("application"), "domain", []string{"model"}, "application"}, {nil, "domain", []string{"model"}, "domain"},
				{nil, "", []string{"model"}, "model"}, {nil, "", nil, "teaql"}, {nil, "", []string{"one", "two"}, "teaql"},
			} {
				directory := filepath.Join(t.TempDir(), strconv.Itoa(index))
				env := map[string]string{"TEAQL_DOMAIN": test.domain, "TEAQL_LOG_FORMAT": format}
				logger, err := NewRuntimeLogger(RuntimeLoggingOptions{LogName: test.explicit, ModelNames: test.models, Environment: env, WorkingDirectory: directory})
				if err != nil {
					t.Fatal(err)
				}
				if logger.LogName != test.name || logger.LogFormat != format || logger.Path != filepath.Join(directory, "logs", test.name+".log") {
					t.Fatal("selection", logger)
				}
				env["TEAQL_LOG_FORMAT"] = "invalid-unused"
				env["TEAQL_DOMAIN"] = "changed"
				context := NewUserContextWithLogger(logger)
				_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
				scope.Success(nil)
				if logger.LogFormat != format {
					t.Fatal("mutable format")
				}
				_, err = NewRuntimeLogger(RuntimeLoggingOptions{LogName: loggingName(test.name), LogFormat: loggingName(format), Environment: map[string]string{}, WorkingDirectory: directory})
				if err != nil {
					t.Fatal(err)
				}
				if len(presentationRead(t, format, logger.Path)) != 3 {
					t.Fatal("two starts did not append")
				}
				logger, err = NewRuntimeLogger(RuntimeLoggingOptions{LogFormat: loggingName(format), Environment: map[string]string{"TEAQL_LOG_FORMAT": "invalid-unused", "TEAQL_LOG_ENDPOINT": "override/custom.log"}, WorkingDirectory: directory})
				if err != nil {
					t.Fatal(err)
				}
				if logger.Path != filepath.Join(directory, "override", "custom.log") {
					t.Fatal("endpoint changed by format")
				}
			}
		})
	}
	logger := newTestLogger(t, t.TempDir(), map[string]string{})
	if logger.LogFormat != "json" {
		t.Fatal("default changed")
	}
	for _, value := range []string{"", "debug", "JSON", "human\nFORMAT-CANARY"} {
		for _, app := range []bool{false, true} {
			options := RuntimeLoggingOptions{Environment: map[string]string{}, WorkingDirectory: t.TempDir()}
			if app {
				options.LogFormat = &value
			} else {
				options.Environment["TEAQL_LOG_FORMAT"] = value
			}
			if _, err := NewRuntimeLogger(options); err == nil || strings.Contains(err.Error(), "CANARY") {
				t.Fatal("invalid format accepted/leaked")
			}
		}
	}
}

func TestRuntimeLoggingPresentationsStreamsAndEquivalentTypedEvent(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		for _, endpoint := range []string{"stdout", "stderr"} {
			t.Run(format+endpoint, func(t *testing.T) {
				reader, writer, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				old := os.Stdout
				if endpoint == "stderr" {
					old = os.Stderr
					os.Stderr = writer
				} else {
					os.Stdout = writer
				}
				defer func() {
					if endpoint == "stderr" {
						os.Stderr = old
					} else {
						os.Stdout = old
					}
					reader.Close()
					writer.Close()
				}()
				logger := newTestLogger(t, t.TempDir(), map[string]string{"TEAQL_LOG_FORMAT": format, "TEAQL_LOG_ENDPOINT": endpoint})
				fields := map[string]interface{}{"category": "SQL", "trace_id": "0123456789abcdef0123456789abcdef", "outcome": "retrying", "error_code": "connection_timeout", "entity_id": int64(9007199254740993), "enabled": true, "result_count": 0, "omission_reason": nil, "changes": []string{"***"}, "sql": "SELECT id\nFROM customer\n\x1b\r\u0085\u2028"}
				message := "Readable \\u001b\nFORGED\x1b\u0085\u2029"
				if !logger.Write("WARN", "unknown.kind", message, fields) {
					t.Fatal("write")
				}
				writer.Close()
				data, err := io.ReadAll(reader)
				if err != nil {
					t.Fatal(err)
				}
				rows := presentationRecords(t, format, data)
				if len(rows) != 2 || rows[1]["level"] != "WARN" || rows[1]["message"] != message {
					t.Fatal("event changed", rows)
				}
				delete(rows[1], "timestamp")
				for k, v := range map[string]interface{}{"level": "WARN", "log_name": "teaql", "runtime": "go", "event": "unknown.kind", "message": message} {
					fields[k] = v
				}
				want, _ := json.Marshal(fields)
				got, _ := json.Marshal(rows[1])
				if string(want) != string(got) {
					t.Fatalf("typed fields changed: %s != %s", got, want)
				}
			})
		}
	}
}

func TestRuntimeLoggingPresentationsNativePrivacyLevelsAndCorrelation(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		t.Run(format, func(t *testing.T) {
			logger := newTestLogger(t, t.TempDir(), map[string]string{"TEAQL_LOG_FORMAT": format})
			context := NewUserContextWithLogger(logger).WithRuntimeTelemetry(loggingCorrelationTelemetry{})
			_, scope := StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
			scope.Success(nil)
			for _, original := range []error{&RuntimeError{Type: "MutationPolicy", Message: "POLICY-CANARY"}, errors.New("DRIVER-CANARY")} {
				_, scope = StartRuntimeOperation(context, context.RuntimeTelemetry(), NewRuntimeOperation("query", "probe", nil))
				FailRuntimeOperation(scope, original)
			}
			for _, ack := range []string{"", "true", plaintextLogAck} {
				t.Setenv(plaintextLogEnv, ack)
				secret := "PASSWORD-CANARY"
				context.RecordExecutionMetadata(data_service.ExecutionMetadata{Operation: data_service.OpQuery, ExecutionOutcome: "retrying", DebugQuery: &secret,
					SQLCorrelation: &data_service.SQLExecutionCorrelation{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef"}})
				event := Created("Customer", core.Record{"id": core.ValI64(42), "password": core.ValText(secret)})
				if err := context.SendEvent(event); err != nil {
					t.Fatal(err)
				}
				if event.Values["password"].V != secret || event.Values["id"].V != int64(42) {
					t.Fatal("business inputs mutated")
				}
			}
			rows := presentationRead(t, format, logger.Path)
			for i, want := range []string{"INFO", "INFO", "WARN", "ERROR"} {
				if rows[i]["level"] != want {
					t.Fatal("severity changed", rows)
				}
			}
			for _, row := range rows {
				if row["category"] == "SQL" && (row["level"] != "WARN" || row["trace_id"] != "0123456789abcdef0123456789abcdef") {
					t.Fatal("SQL classification/correlation")
				}
			}
			data, _ := os.ReadFile(logger.Path)
			if strings.Contains(string(data), "CANARY") {
				t.Fatal("raw private value reached renderer")
			}
		})
	}
}

func TestRuntimeLoggingPresentationsPositiveRetentionOrder(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		t.Run(format, func(t *testing.T) {
			logger := newTestLogger(t, t.TempDir(), map[string]string{"TEAQL_LOG_FORMAT": format, "TEAQL_LOG_MAX_SIZE": "256", "TEAQL_LOG_MAX_FILES": "2"})
			for i := 0; i < 5; i++ {
				if !logger.Write("INFO", "retention", strings.Repeat("界", 200), map[string]interface{}{"sequence": i, "sql": "SELECT\n" + strconv.Itoa(i)}) {
					t.Fatal("write")
				}
			}
			for suffix, expected := range map[string]int{"": 4, ".1": 3, ".2": 2} {
				rows := presentationRead(t, format, logger.Path+suffix)
				if len(rows) != 1 || rows[0]["sequence"] != json.Number(strconv.Itoa(expected)) || rows[0]["sql"] != "SELECT\n"+strconv.Itoa(expected) {
					t.Fatal("numeric archive order or split event", rows)
				}
			}
			paths, _ := filepath.Glob(logger.Path + "*")
			if len(paths) != 3 {
				t.Fatal("positive retention limit", paths)
			}
		})
	}
}

func TestRuntimeLoggingPresentationsRotationsAndWholeBlocks(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		for _, override := range []bool{false, true} {
			t.Run(fmt.Sprint(format, override), func(t *testing.T) {
				env := map[string]string{"TEAQL_LOG_FORMAT": format, "TEAQL_LOG_MAX_SIZE": "1K", "TEAQL_LOG_MAX_FILES": "100"}
				if override {
					env["TEAQL_LOG_ENDPOINT"] = "override/custom.log"
				}
				directory := t.TempDir()
				logger := newTestLogger(t, directory, env)
				for index := 0; index < 40; index++ {
					if !logger.Write("INFO", "regular", strings.Repeat("界", 25), map[string]interface{}{"index": index, "sql": "SELECT id\nFROM customer"}) {
						t.Fatal("write")
					}
				}
				paths, _ := filepath.Glob(logger.Path + "*")
				rows := presentationRead(t, format, paths...)
				seen := map[json.Number]bool{}
				for _, p := range paths {
					info, _ := os.Stat(p)
					if info.Size() > 1024 || info.Size() == 0 {
						t.Fatal("UTF8 byte limit", p)
					}
				}
				for _, row := range rows {
					if row["event"] == "regular" {
						if row["sql"] != "SELECT id\nFROM customer" {
							t.Fatal("split SQL")
						}
						seen[row["index"].(json.Number)] = true
					}
				}
				if len(rows) != 41 || len(seen) != 40 {
					t.Fatal("lost events")
				}
				if !logger.Write("WARN", "oversize", "whole", map[string]interface{}{"sql": strings.Repeat("界", 2000) + "\nLAST"}) {
					t.Fatal("oversize")
				}
				before, _ := os.ReadFile(logger.Path)
				if len(presentationRecords(t, format, before)) != 1 {
					t.Fatal("split oversized block")
				}
				logger.Write("INFO", "after", "next", nil)
				archived, _ := os.ReadFile(logger.Path + ".1")
				if string(archived) != string(before) {
					t.Fatal("changed oversized block")
				}
				yesterday := time.Now().Add(-48 * time.Hour)
				if err := os.Chtimes(logger.Path, yesterday, yesterday); err != nil {
					t.Fatal(err)
				}
				logger.Write("INFO", "day", "UTC rotation", nil)
				if presentationRead(t, format, logger.Path+".1")[0]["event"] != "after" {
					t.Fatal("day rotation")
				}
				os.Chtimes(logger.Path, yesterday, yesterday)
				logger = newTestLogger(t, directory, env)
				if presentationRead(t, format, logger.Path+".1")[0]["event"] != "day" {
					t.Fatal("restart rotation")
				}
				env["TEAQL_LOG_MAX_FILES"] = "0"
				logger = newTestLogger(t, directory, env)
				logger.Write("INFO", "last", strings.Repeat("界", 2000), nil)
				paths, _ = filepath.Glob(logger.Path + "*")
				if len(paths) != 1 || len(presentationRead(t, format, paths...)) != 1 {
					t.Fatal("zero retention")
				}
			})
		}
	}
}

func TestRuntimeLoggingPresentationsConcurrentThreads(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			var workers sync.WaitGroup
			var logger *RuntimeLogger
			for w := 0; w < 4; w++ {
				logger = newTestLogger(t, dir, map[string]string{"TEAQL_LOG_FORMAT": format, "TEAQL_LOG_MAX_SIZE": "2K", "TEAQL_LOG_MAX_FILES": "100"})
				workers.Add(1)
				go func(id int, l *RuntimeLogger) {
					defer workers.Done()
					for i := 0; i < 30; i++ {
						key := fmt.Sprintf("%d/%d", id, i)
						if !l.Write("INFO", "parallel", "多线程", map[string]interface{}{"id": key, "sql": "SELECT\n" + key}) {
							t.Error("write failed")
						}
					}
				}(w, logger)
			}
			workers.Wait()
			assertPresentationIDs(t, format, logger.Path, 120, 4)
		})
	}
}

func assertPresentationIDs(t *testing.T, format, path string, count, starts int) {
	t.Helper()
	paths, _ := filepath.Glob(path + "*")
	rows := presentationRead(t, format, paths...)
	ids := map[string]bool{}
	for _, row := range rows {
		if id, ok := row["id"].(string); ok {
			if ids[id] || row["sql"] != "SELECT\n"+id {
				t.Fatal("interleaved/duplicate block", row)
			}
			ids[id] = true
		}
	}
	if len(ids) != count || len(rows) != count+starts {
		t.Fatal("lost blocks", len(ids), len(rows))
	}
}

func TestRuntimeLoggingPresentationWorker(t *testing.T) {
	if os.Getenv("TEAQL_PRESENTATION_CHILD") != "yes" {
		return
	}
	logger := newTestLogger(t, os.Getenv("TEAQL_PRESENTATION_DIRECTORY"), map[string]string{"TEAQL_LOG_FORMAT": os.Getenv("TEAQL_PRESENTATION_FORMAT"), "TEAQL_LOG_MAX_SIZE": "2K", "TEAQL_LOG_MAX_FILES": "100"})
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("%s/%d", os.Getenv("TEAQL_PRESENTATION_WORKER"), i)
		if !logger.Write("INFO", "parallel", "原生并发", map[string]interface{}{"id": id, "sql": "SELECT\n" + id}) {
			t.Fatal("child write")
		}
	}
}

func TestRuntimeLoggingPresentationsSeparateProcessesAndRestart(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			var commands []*exec.Cmd
			t.Cleanup(func() {
				for _, c := range commands {
					if c.ProcessState == nil {
						_ = c.Process.Kill()
						_ = c.Wait()
					}
				}
			})
			for worker := 0; worker < 5; worker++ {
				command := exec.Command(os.Args[0], "-test.run=^TestRuntimeLoggingPresentationWorker$", "-test.timeout=30s")
				command.Env = []string{"TEAQL_PRESENTATION_CHILD=yes", "TEAQL_PRESENTATION_DIRECTORY=" + dir, "TEAQL_PRESENTATION_FORMAT=" + format, fmt.Sprintf("TEAQL_PRESENTATION_WORKER=%d", worker), "TEAQL_TELEMETRY_ENABLED=false"}
				command.Stdout, command.Stderr = os.Stdout, os.Stderr
				if err := command.Start(); err != nil {
					t.Fatal(err)
				}
				commands = append(commands, command)
				if worker == 3 {
					for _, c := range commands {
						if err := c.Wait(); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			if err := commands[4].Wait(); err != nil {
				t.Fatal(err)
			}
			assertPresentationIDs(t, format, filepath.Join(dir, "logs", "teaql.log"), 150, 5)
		})
	}
}

func TestRuntimeLoggingPresentationsMixedFormatAndIOPolicy(t *testing.T) {
	for _, format := range []string{"json", "human"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			logger := newTestLogger(t, dir, map[string]string{"TEAQL_LOG_FORMAT": format})
			other := "human"
			if format == "human" {
				other = "json"
			}
			if err := os.WriteFile(logger.Path+".10", []byte("preserve archive bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			snapshot := func() map[string]string {
				paths, _ := filepath.Glob(logger.Path + "*")
				files := map[string]string{}
				for _, p := range paths {
					b, e := os.ReadFile(p)
					if e != nil {
						t.Fatal(e)
					}
					files[p] = string(b)
				}
				return files
			}
			before := snapshot()
			_, err := NewRuntimeLogger(RuntimeLoggingOptions{WorkingDirectory: dir, Environment: map[string]string{"TEAQL_LOG_FORMAT": other, "TEAQL_LOG_MAX_SIZE": "1", "TEAQL_LOG_MAX_FILES": "0"}})
			if err == nil || !reflect.DeepEqual(before, snapshot()) {
				t.Fatal("mixed startup changed files")
			}
			replacement := newTestLogger(t, t.TempDir(), map[string]string{"TEAQL_LOG_FORMAT": other})
			data, _ := os.ReadFile(replacement.Path)
			if err := os.WriteFile(logger.Path, data, 0600); err != nil {
				t.Fatal(err)
			}
			before = snapshot()
			if logger.Write("INFO", "later", "business completion", nil) || !reflect.DeepEqual(before, snapshot()) {
				t.Fatal("mixed append changed files")
			}
			if err := os.Remove(logger.Path); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(logger.Path, 0700); err != nil {
				t.Fatal(err)
			}
			if logger.Write("ERROR", "later", "I/O failure", nil) {
				t.Fatal("I/O failure accepted")
			}
			_, err = NewRuntimeLogger(RuntimeLoggingOptions{WorkingDirectory: dir, LogFormat: &format, Environment: map[string]string{}})
			if err == nil {
				t.Fatal("I/O startup must fail")
			}
		})
	}
}
