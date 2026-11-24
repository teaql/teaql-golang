module checker-fix-advanced-core-workspace

go 1.18

require (
	github.com/teaql/teaql-golang v0.6.0
	checker-fix-advanced-core-workspace/lib v0.0.0
)

replace checker-fix-advanced-core-workspace/lib => ../../fix-e-alignment/golang-lib-core
replace github.com/teaql/teaql-golang => ../../..
