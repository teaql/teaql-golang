module example.com/task_board

go 1.18

replace robot-kanban-service-core-workspace/lib => ./generated

replace github.com/teaql/teaql-golang => ../../

require (
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/teaql/teaql-golang v0.0.0
	robot-kanban-service-core-workspace/lib v0.0.0-00010101000000-000000000000
)

require github.com/shopspring/decimal v1.4.0 // indirect
