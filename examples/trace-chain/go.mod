module teaql-trace-chain-example

go 1.18

require (
	github.com/teaql/teaql-golang v0.2.9
	trace-chain-service-core-workspace/lib v0.0.0
)

replace github.com/teaql/teaql-golang => ../..

replace trace-chain-service-core-workspace/lib => ./lib
