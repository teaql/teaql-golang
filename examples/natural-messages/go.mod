module teaql-natural-messages-example

go 1.18

require (
    natural-message-service-core-workspace/lib v0.0.0
    github.com/teaql/teaql-golang v0.5.0-dev.0
)

replace natural-message-service-core-workspace/lib => ./fixture
replace github.com/teaql/teaql-golang => ../..
