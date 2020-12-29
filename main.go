package main

import (
	"github.com/pga2rn/ib-dtm_framework/service"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:   "framework",
		Action: service.Entry,
		Before: service.Init,
		After:  service.Done,
	}

	if err := app.Run(os.Args); err != nil {
		logutil.LoggerList["main"].Fatalf("failed to start the application")
	}
}
