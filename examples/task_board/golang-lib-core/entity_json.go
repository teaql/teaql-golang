package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"robot-kanban-service-core-workspace/lib/internal/generatedjson"
	"robot-kanban-service-core-workspace/lib/platform"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Task Status": return task_status.TeaqlTaskStatusFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Task": return task.TeaqlTaskFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Task Execution Log": return task_execution_log.TeaqlTaskExecutionLogFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// PlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*platform.Platform, error) {
	record, err := generatedjson.Model.Parse("Platform", payload, profile...)
	if err != nil { return nil, err }
	return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// TaskStatusFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func TaskStatusFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*task_status.TaskStatus, error) {
	record, err := generatedjson.Model.Parse("Task Status", payload, profile...)
	if err != nil { return nil, err }
	return task_status.TeaqlTaskStatusFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// TaskFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func TaskFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*task.Task, error) {
	record, err := generatedjson.Model.Parse("Task", payload, profile...)
	if err != nil { return nil, err }
	return task.TeaqlTaskFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// TaskExecutionLogFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func TaskExecutionLogFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*task_execution_log.TaskExecutionLog, error) {
	record, err := generatedjson.Model.Parse("Task Execution Log", payload, profile...)
	if err != nil { return nil, err }
	return task_execution_log.TeaqlTaskExecutionLogFromJSONRecord(record, teaqlEntityFromJSONRecord)
}