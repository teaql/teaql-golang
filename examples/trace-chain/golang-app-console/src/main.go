package main

import (
	"fmt"
	"os"
	src "trace-chain-service-core-workspace/lib"
)

func main() {
		// SQLite can use a local default; server providers require an explicit URL.
		if os.Getenv("TRACE_CHAIN_SERVICE_CORE_DATABASE_URL") == "" {
			os.Setenv("TRACE_CHAIN_SERVICE_CORE_DATABASE_URL", "./teaql.db")
		}
	context, err := src.ServiceRuntimeFromEnv()
	if err != nil {
		panic(err)
	}
	if err := src.EnsureSchema(context); err != nil {
		panic(err)
	}
	fmt.Println("trace-chain-service TeaQL console ready")
}
