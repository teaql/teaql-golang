module school-management-service-core-workspace

go 1.18

require (
	github.com/teaql/teaql-golang v0.6.0
	school-management-service-core-workspace/lib v0.0.0
)

replace school-management-service-core-workspace/lib => ../golang-lib-core

replace github.com/teaql/teaql-golang => ../../..
