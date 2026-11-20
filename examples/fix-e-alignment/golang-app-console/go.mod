module teaql-fix-e-example

go 1.18

require (
    checker-fix-advanced-core-workspace/lib v0.0.0
    github.com/teaql/teaql-golang v0.6.0
)

replace checker-fix-advanced-core-workspace/lib => ../golang-lib-core
replace github.com/teaql/teaql-golang => ../../..
