module robot-kanban-service-core-workspace

go 1.18

require (
	github.com/teaql/teaql-golang v0.6.0
	robot-kanban-service-core-workspace/lib v0.0.0
)

replace robot-kanban-service-core-workspace/lib => ../golang-lib-core

replace github.com/teaql/teaql-golang => ../../..
