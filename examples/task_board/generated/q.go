package lib

import (
	"robot-kanban-service-core-workspace/lib/src/platform"
	"robot-kanban-service-core-workspace/lib/src/task_status"
	"robot-kanban-service-core-workspace/lib/src/task"
	"robot-kanban-service-core-workspace/lib/src/task_execution_log"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) TaskStatuss() *task_status.TaskStatusRequest {
	return task_status.NewTaskStatusRequest()
}

func (q *QType) TaskStatussMinimal() *task_status.TaskStatusRequest {
	return task_status.NewTaskStatusRequest()
}

func (q *QType) Tasks() *task.TaskRequest {
	return task.NewTaskRequest()
}

func (q *QType) TasksMinimal() *task.TaskRequest {
	return task.NewTaskRequest()
}

func (q *QType) TaskExecutionLogs() *task_execution_log.TaskExecutionLogRequest {
	return task_execution_log.NewTaskExecutionLogRequest()
}

func (q *QType) TaskExecutionLogsMinimal() *task_execution_log.TaskExecutionLogRequest {
	return task_execution_log.NewTaskExecutionLogRequest()
}