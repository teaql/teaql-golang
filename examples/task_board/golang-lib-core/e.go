package lib

import (
	"robot-kanban-service-core-workspace/lib/platform"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) TaskStatus(value *task_status.TaskStatus) *task_status.TaskStatusExpression {
	return task_status.NewTaskStatusExpression(value)
}

func (expressionFacade) Task(value *task.Task) *task.TaskExpression {
	return task.NewTaskExpression(value)
}

func (expressionFacade) TaskExecutionLog(value *task_execution_log.TaskExecutionLog) *task_execution_log.TaskExecutionLogExpression {
	return task_execution_log.NewTaskExecutionLogExpression(value)
}
