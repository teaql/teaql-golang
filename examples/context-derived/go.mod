module teaql-context-derived-example

go 1.18

require (
    context-fix-service-core-workspace/lib v0.0.0
    github.com/teaql/teaql-golang v0.4.0
)

replace context-fix-service-core-workspace/lib => ./fixture
replace github.com/teaql/teaql-golang => ../..
