package main

import (
	"fmt"
	"os"
	src "checker-fix-advanced-core-workspace/lib"
)

func main() {
		// SQLite can use a local default; server providers require an explicit URL.
		if os.Getenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL") == "" {
			os.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", "./teaql.db")
		}
	context, err := src.ServiceRuntimeFromEnv()
	if err != nil {
		panic(err)
	}
	if err := src.EnsureSchema(context); err != nil {
		panic(err)
	}
	fmt.Println("checker-fix-advanced TeaQL console ready")
}
