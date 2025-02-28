package main

import (
	"github.com/pga2rn/ib-dtm_framework/service"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/urfave/cli/v2"
	"os"
	runtimeDebug "runtime/debug"
)

var PackageName = "main"

func main() {
	app := &cli.App{
		Name:   "framework",
		Action: service.Entry,
		Before: service.Init,
		After:  service.Done,
	}

	defer func() {
		if x := recover(); x != nil {
			logutil.GetLogger(PackageName).Errorf("Runtime panic: %v\n%v", x, string(runtimeDebug.Stack()))
			panic(x)
		}
	}()

	if err := app.Run(os.Args); err != nil {
		logutil.GetLogger(PackageName).Fatalf("failed to start the application")
	}
}
