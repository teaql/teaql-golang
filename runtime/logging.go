package runtime

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

// RuntimeLoggingOptions describes application startup, not a request. A nil
// Environment reads only logging settings from the process.
type RuntimeLoggingOptions struct {
	LogName          *string
	LogFormat        *string
	ModelNames       []string
	Environment      map[string]string
	WorkingDirectory string
}

// RuntimeLogger owns the resolved destination. Write accepts already projected
// fields; SQL and audit callers must use the existing privacy projection first.
type RuntimeLogger struct {
	LogName   string
	LogFormat string
	Path      string
	MaxSize   int64
	MaxFiles  int
	stream    string
	mu        *sync.Mutex
}

var loggingSettings = []string{"TEAQL_DOMAIN", "TEAQL_LOG_ENDPOINT", "TEAQL_LOG_MAX_SIZE", "TEAQL_LOG_MAX_FILES", "TEAQL_LOG_FORMAT"}
var loggingDefaults sync.Map
var loggingDefaultMu sync.Mutex
var loggingLocks sync.Map
var loggingSize = regexp.MustCompile(`^([0-9]+)\s*(K|KB|M|MB|G|GB)?$`)
var loggingArchive = regexp.MustCompile(`^[1-9][0-9]*$`)
var loggingDevice = regexp.MustCompile(`(?i)^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$`)
var loggingHumanHeader = regexp.MustCompile(`^\d{4}-\d\d-\d\dT\S+(Z|\+00:00) \[TeaQL\]\[.*\]\[(INFO|WARN|ERROR)\] .+ runtime=\S+$`)

func loggingEnvironment() map[string]string {
	values := make(map[string]string)
	for _, key := range loggingSettings {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}
	return values
}

func loggingFailure(name, code string) {
	data, _ := json.Marshal(map[string]interface{}{"timestamp": time.Now().UTC().Format(time.RFC3339Nano), "level": "ERROR", "log_name": name, "runtime": "go", "event": "logging.failure", "message": "Runtime logging failed", "error_code": code})
	_, _ = os.Stderr.Write(append(data, '\n'))
}

func validLogName(name string) bool {
	if name == "" || !utf8.ValidString(name) || name == "." || name == ".." || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") || loggingDevice.MatchString(strings.Split(name, ".")[0]) {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) || strings.ContainsRune("<>:\"/\\|?*", r) {
			return false
		}
	}
	return true
}

func parseLogSize(value string) (int64, error) {
	match := loggingSize.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(value)))
	if match == nil {
		return 0, errors.New("invalid_log_max_size")
	}
	amount, err := strconv.ParseInt(match[1], 10, 64)
	multiplier := int64(1)
	switch match[2] {
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	}
	if err != nil || amount <= 0 || amount > (1<<63-1)/multiplier {
		return 0, errors.New("invalid_log_max_size")
	}
	return amount * multiplier, nil
}

func NewRuntimeLogger(options RuntimeLoggingOptions) (logger *RuntimeLogger, err error) {
	name := "teaql"
	defer func() {
		if err != nil {
			loggingFailure(name, "log_destination_unavailable")
			logger, err = nil, errors.New("runtime logging initialization failed")
		}
	}()
	environment := options.Environment
	if environment == nil {
		environment = loggingEnvironment()
	}
	if options.LogName != nil {
		name = *options.LogName
	} else if environment["TEAQL_DOMAIN"] != "" {
		name = environment["TEAQL_DOMAIN"]
	} else if len(options.ModelNames) == 1 {
		name = options.ModelNames[0]
	}
	if !validLogName(name) {
		name = "teaql"
		return nil, errors.New("invalid_log_name")
	}
	format := "json"
	if value, exists := environment["TEAQL_LOG_FORMAT"]; exists {
		format = value
	}
	if options.LogFormat != nil {
		format = *options.LogFormat
	}
	if format != "json" && format != "human" {
		return nil, errors.New("invalid_log_format")
	}
	logger = &RuntimeLogger{LogName: name, LogFormat: format, MaxSize: 50 * 1024 * 1024, MaxFiles: 7, mu: &sync.Mutex{}}
	if value, exists := environment["TEAQL_LOG_MAX_SIZE"]; exists {
		logger.MaxSize, err = parseLogSize(value)
		if err != nil {
			return nil, err
		}
	}
	if value, exists := environment["TEAQL_LOG_MAX_FILES"]; exists {
		files, parseErr := strconv.ParseUint(value, 10, 31)
		if parseErr != nil || value == "" {
			return nil, errors.New("invalid_log_max_files")
		}
		logger.MaxFiles = int(files)
	}
	endpoint := environment["TEAQL_LOG_ENDPOINT"]
	if endpoint == "" {
		endpoint = filepath.Join("logs", name+".log")
	}
	if endpoint == "stdout" || endpoint == "stderr" || endpoint == "-" {
		logger.stream = endpoint
	} else {
		if strings.HasPrefix(endpoint, "file:") {
			parsed, parseErr := url.Parse(endpoint)
			if parseErr != nil || (parsed.Host != "" && parsed.Host != "localhost") || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path == "" {
				return nil, errors.New("invalid_log_endpoint")
			}
			endpoint = parsed.Path
		} else if strings.Contains(endpoint, "://") {
			return nil, errors.New("invalid_log_endpoint")
		}
		base := options.WorkingDirectory
		if base == "" {
			base, err = os.Getwd()
			if err != nil {
				return nil, err
			}
		}
		if !filepath.IsAbs(endpoint) {
			endpoint = filepath.Join(base, endpoint)
		}
		logger.Path, err = filepath.Abs(endpoint)
		if err != nil {
			return nil, err
		}
		if err = os.MkdirAll(filepath.Dir(logger.Path), 0700); err != nil {
			return nil, err
		}
		parent, resolveErr := filepath.EvalSymlinks(filepath.Dir(logger.Path))
		if resolveErr != nil {
			return nil, resolveErr
		}
		logger.Path = filepath.Join(parent, filepath.Base(logger.Path))
		if resolved, resolveErr := filepath.EvalSymlinks(logger.Path); resolveErr == nil {
			logger.Path = resolved
		}
		mutex, _ := loggingLocks.LoadOrStore(logger.Path, &sync.Mutex{})
		logger.mu = mutex.(*sync.Mutex)
	}
	err = logger.emit(logger.record("INFO", "runtime.start", "TeaQL runtime started", nil))
	return logger, err
}

func defaultRuntimeLogger() *RuntimeLogger {
	environment := loggingEnvironment()
	cwd, err := os.Getwd()
	if err != nil {
		loggingFailure("teaql", "log_destination_unavailable")
		panic("runtime logging initialization failed")
	}
	key, _ := json.Marshal([]interface{}{cwd, environment})
	loggingDefaultMu.Lock()
	defer loggingDefaultMu.Unlock()
	if logger, ok := loggingDefaults.Load(string(key)); ok {
		return logger.(*RuntimeLogger)
	}
	logger, err := NewRuntimeLogger(RuntimeLoggingOptions{Environment: environment, WorkingDirectory: cwd})
	if err != nil {
		panic(err)
	}
	loggingDefaults.Store(string(key), logger)
	return logger
}

func (l *RuntimeLogger) record(level, event, message string, fields map[string]interface{}) map[string]interface{} {
	record := make(map[string]interface{}, len(fields)+6)
	for key, value := range fields {
		record[key] = value
	}
	record["timestamp"], record["level"], record["log_name"] = time.Now().UTC().Format(time.RFC3339Nano), level, l.LogName
	record["runtime"], record["event"], record["message"] = "go", event, message
	return record
}

func (l *RuntimeLogger) Write(level, event, message string, safeFields map[string]interface{}) (ok bool) {
	defer func() {
		if recover() != nil {
			loggingFailure(l.LogName, "log_write_failed")
			ok = false
		}
	}()
	if level != "INFO" && level != "WARN" && level != "ERROR" {
		loggingFailure(l.LogName, "invalid_log_level")
		return false
	}
	if err := l.emit(l.record(level, event, message, safeFields)); err != nil {
		loggingFailure(l.LogName, "log_write_failed")
		return false
	}
	return true
}

func (l *RuntimeLogger) archives() ([]int, error) {
	entries, err := os.ReadDir(filepath.Dir(l.Path))
	if err != nil {
		return nil, err
	}
	var numbers []int
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), filepath.Base(l.Path)+".") {
			continue
		}
		suffix := strings.TrimPrefix(entry.Name(), filepath.Base(l.Path)+".")
		if !loggingArchive.MatchString(suffix) {
			continue
		}
		number, parseErr := strconv.Atoi(suffix)
		if parseErr != nil || !entry.Type().IsRegular() {
			return nil, errors.New("invalid_log_archive")
		}
		numbers = append(numbers, number)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(numbers)))
	return numbers, nil
}

func (l *RuntimeLogger) emit(record map[string]interface{}) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if l.LogFormat == "human" {
		data, err = humanLogRecord(data)
		if err != nil {
			return err
		}
	} else {
		data = append(data, '\n')
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.stream != "" {
		var writer io.Writer = os.Stdout
		if l.stream == "stderr" {
			writer = os.Stderr
		}
		_, err = writer.Write(data)
		return err
	}
	// Atomic exclusive creation is shared by all runtime implementations. The
	// lock covers rotation AND append; never retain an active-file descriptor.
	deadline := time.Now().Add(30 * time.Second)
	var lock *os.File
	for {
		lock, err = os.OpenFile(l.Path+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			break
		}
		if !os.IsExist(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
	defer func() { _ = lock.Close(); _ = os.Remove(l.Path + ".lock") }()
	// Compatibility precedes every retention or rotation mutation.
	if err = l.checkPresentation(); err != nil {
		return err
	}
	numbers, err := l.archives()
	if err != nil {
		return err
	}
	for _, number := range numbers {
		if number > l.MaxFiles {
			if err = os.Remove(l.Path + "." + strconv.Itoa(number)); err != nil {
				return err
			}
		}
	}
	info, err := os.Stat(l.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if info != nil {
		if !info.Mode().IsRegular() {
			return errors.New("log_destination_not_file")
		}
		if info.Size() > 0 && (int64(len(data)) > l.MaxSize-info.Size() || info.ModTime().UTC().Format("2006-01-02") != time.Now().UTC().Format("2006-01-02")) {
			for _, number := range numbers {
				if number > l.MaxFiles {
					continue
				}
				path := l.Path + "." + strconv.Itoa(number)
				if number == l.MaxFiles {
					err = os.Remove(path)
				} else {
					err = os.Rename(path, l.Path+"."+strconv.Itoa(number+1))
				}
				if err != nil {
					return err
				}
			}
			if l.MaxFiles == 0 {
				err = os.Remove(l.Path)
			} else {
				err = os.Rename(l.Path, l.Path+".1")
			}
			if err != nil {
				return err
			}
		}
	}
	file, err := os.OpenFile(l.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func humanLogText(value string, escapeBackslash bool) string {
	var output strings.Builder
	for _, r := range value {
		if escapeBackslash && r == '\\' {
			output.WriteString(`\\`)
		} else if r < 32 || (r >= 127 && r <= 159) || r == '\u2028' || r == '\u2029' {
			fmt.Fprintf(&output, `\u%04x`, r)
		} else {
			output.WriteRune(r)
		}
	}
	return output.String()
}

func humanLogRecord(encoded []byte) ([]byte, error) {
	// UseNumber keeps large typed IDs exact in the event shared by both modes.
	var record map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	if err := decoder.Decode(&record); err != nil {
		return nil, err
	}
	text := func(key string) string { return humanLogText(record[key].(string), true) }
	var output strings.Builder
	fmt.Fprintf(&output, "%s [TeaQL][%s][%s] %s runtime=%s\n  %s\n", text("timestamp"), text("log_name"), text("level"), text("event"), text("runtime"), text("message"))
	var keys []string
	for key := range record {
		switch key {
		case "timestamp", "level", "log_name", "runtime", "event", "message":
		default:
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		if sql, ok := record[key].(string); key == "sql" && ok {
			output.WriteString("  sql:\n")
			for _, line := range strings.Split(sql, "\n") {
				output.WriteString("    " + humanLogText(line, true) + "\n")
			}
		} else {
			value, err := json.Marshal(record[key])
			if err != nil {
				return nil, err
			}
			output.WriteString("  " + humanLogText(key, true) + "=" + humanLogText(string(value), false) + "\n")
		}
	}
	return []byte(output.String()), nil
}

func (l *RuntimeLogger) checkPresentation() error {
	info, err := os.Stat(l.Path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("log_destination_not_file")
	}
	if info.Size() == 0 {
		return nil
	}
	file, err := os.Open(l.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	line, err := bufio.NewReader(file).ReadString('\n')
	if err != nil {
		return errors.New("incompatible_log_format")
	}
	line = strings.TrimSuffix(line, "\n")
	if l.LogFormat == "human" && loggingHumanHeader.MatchString(line) {
		return nil
	}
	if l.LogFormat == "json" {
		var record map[string]interface{}
		if json.Unmarshal([]byte(line), &record) == nil {
			valid := true
			for _, key := range []string{"timestamp", "level", "log_name", "runtime", "event", "message"} {
				if _, ok := record[key].(string); !ok {
					valid = false
				}
			}
			if valid && (record["level"] == "INFO" || record["level"] == "WARN" || record["level"] == "ERROR") {
				return nil
			}
		}
	}
	return errors.New("incompatible_log_format")
}

type runtimeSQLLogSink struct{ logger *RuntimeLogger }

func (s runtimeSQLLogSink) WriteSQLLog(metadata data_service.ExecutionMetadata) {
	safe := projectedSQLMetadata(metadata, false)
	fields := map[string]interface{}{"category": "SQL", "outcome": safe.ExecutionOutcome,
		"sql": safe.DebugQuery, "omission_reason": safe.OmissionReason,
		"comment": safe.Comment, "purpose": safe.Purpose, "audit_reason": safe.AuditReason,
		"trace_chain": safe.RequestTraceChain, "trace_path": safe.TraceChain, "mutation_lineage": safe.MutationLineage,
		"elapsed_us": safe.EndedAt.Sub(safe.StartedAt).Microseconds()}
	if safe.ResultCount != nil {
		fields["result_count"] = *safe.ResultCount
	}
	if safe.AffectedRows != nil {
		fields["affected_rows"] = *safe.AffectedRows
	}
	if safe.SQLCorrelation != nil {
		fields["trace_id"] = safe.SQLCorrelation.TraceID
	}
	level := "INFO"
	if safe.ExecutionOutcome == "failure" || safe.ExecutionOutcome == "error" {
		level = "ERROR"
	} else if safe.ExecutionOutcome == "retrying" {
		level = "WARN"
	}
	s.logger.Write(level, "sql."+safe.ExecutionOutcome, "TeaQL SQL execution", fields)
}

func (c *UserContext) logAuditEvent(event *RawAuditEvent) {
	if c.runtimeLogger == nil {
		return
	}
	var masks []string
	var limit *int
	if descriptor := c.Entity(event.Entity); descriptor != nil {
		for _, field := range descriptor.AuditMaskFlds {
			canonical, err := core.CanonicalLedgerField(descriptor, field)
			if err != nil {
				loggingFailure(c.runtimeLogger.LogName, "log_projection_failed")
				return
			}
			masks = append(masks, canonical)
		}
		limit = descriptor.AuditValueMaxL
	}
	safe := event.BuildSafeEvent(masks, limit)
	c.runtimeLogger.Write("INFO", "audit.event", "TeaQL audit event", map[string]interface{}{
		"category": "AUDIT", "outcome": "success", "entity": safe.Entity, "entity_id": safe.TargetID,
		"fields": safe.Fields, "audit_reason": safe.AuditReason, "trace_chain": safe.TraceChain,
	})
}

type loggedRuntimeScope struct {
	delegate  RuntimeTelemetryScope
	logger    *RuntimeLogger
	operation RuntimeOperation
}

func (s *loggedRuntimeScope) Correlation() *data_service.SQLExecutionCorrelation {
	if correlated, ok := s.delegate.(correlatedTelemetryScope); ok {
		return correlated.Correlation()
	}
	return nil
}
func (s *loggedRuntimeScope) Success(attributes map[string]RuntimeAttributeValue) {
	s.CompleteRuntimeOperation("success", attributes, "")
}
func (s *loggedRuntimeScope) Failure(kind string) {
	s.CompleteRuntimeOperation("failure", nil, kind)
}

func (s *loggedRuntimeScope) failureWithError(err error) {
	category := RuntimeErrorCategory(RuntimeErrorType(err))
	if native, ok := err.(*RuntimeError); ok {
		switch native.Type {
		case "Check":
			category = "validation"
		case "MutationPolicy":
			category = "authorization"
		case "OptimisticLockConflict":
			category = "conflict"
		}
	}
	if _, ok := err.(*core.RequestIntentError); ok {
		category = "validation"
	}
	s.complete("failure", nil, RuntimeErrorType(err), category)
}
func (s *loggedRuntimeScope) CompleteRuntimeOperation(outcome string, attributes map[string]RuntimeAttributeValue, kind string) {
	s.complete(outcome, attributes, kind, RuntimeErrorCategory(kind))
}
func (s *loggedRuntimeScope) complete(outcome string, attributes map[string]RuntimeAttributeValue, kind, category string) {
	level := "INFO"
	fields := map[string]interface{}{"category": strings.ToUpper(s.operation.Family), "outcome": outcome}
	if identity := s.Correlation(); identity != nil && identity.TraceID != "" {
		fields["trace_id"] = identity.TraceID
	}
	if outcome != "success" {
		fields["error_code"] = category
		level = "ERROR"
		if category == "validation" || category == "authorization" || category == "conflict" || outcome == "retrying" {
			level = "WARN"
		}
	}
	s.logger.Write(level, "operation."+outcome, "TeaQL operation "+outcome, fields)
	CompleteRuntimeOperation(s.delegate, outcome, attributes, kind)
}

func (l *RuntimeLogger) String() string { return fmt.Sprintf("RuntimeLogger(%s)", l.LogName) }
