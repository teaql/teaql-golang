package main

import (
	"fmt"
	"os"
	src "context-fix-service-core-workspace/lib"
)

func main() {
		// SQLite can use a local default; server providers require an explicit URL.
		if os.Getenv("CONTEXT_FIX_SERVICE_CORE_DATABASE_URL") == "" {
			os.Setenv("CONTEXT_FIX_SERVICE_CORE_DATABASE_URL", "./teaql.db")
		}
	context, err := src.ServiceRuntimeFromEnv()
	if err != nil {
		panic(err)
	}
	if err := src.EnsureSchema(context); err != nil {
		panic(err)
	}
	fmt.Println("context-fix-service TeaQL console ready")
}
