package lib

import (
	"robot-kanban-service-core-workspace/lib/platform"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) TaskStatuses() *task_status.TaskStatusRequest {
	return task_status.NewTaskStatusRequest()
}

func (q *QType) TaskStatusesMinimal() *task_status.TaskStatusRequest {
	return task_status.NewTaskStatusMinimalRequest()
}

func (q *QType) Tasks() *task.TaskRequest {
	return task.NewTaskRequest()
}

func (q *QType) TasksMinimal() *task.TaskRequest {
	return task.NewTaskMinimalRequest()
}

func (q *QType) TaskExecutionLogs() *task_execution_log.TaskExecutionLogRequest {
	return task_execution_log.NewTaskExecutionLogRequest()
}

func (q *QType) TaskExecutionLogsMinimal() *task_execution_log.TaskExecutionLogRequest {
	return task_execution_log.NewTaskExecutionLogMinimalRequest()
}

func (q *QType) TaskStatusPlatform(entity *task_status.TaskStatus) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}

func (q *QType) TaskStatus(entity *task.Task) (*task_status.TaskStatus, bool) {
	value, ok := entity.RelationEntity("statusEntity")
	if !ok { return nil, false }
	typed, ok := value.(*task_status.TaskStatus)
	return typed, ok
}

func (q *QType) TaskPlatform(entity *task.Task) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}

func (q *QType) TaskExecutionLogTask(entity *task_execution_log.TaskExecutionLog) (*task.Task, bool) {
	value, ok := entity.RelationEntity("taskEntity")
	if !ok { return nil, false }
	typed, ok := value.(*task.Task)
	return typed, ok
}