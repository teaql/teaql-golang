module teaql-sqlite-lifecycle-example

go 1.18

require (
    checker-fix-advanced-core-workspace/lib v0.0.0
    github.com/teaql/teaql-golang v0.3.1
)

replace checker-fix-advanced-core-workspace/lib => ../fix-e-alignment/generated
replace github.com/teaql/teaql-golang => ../..
