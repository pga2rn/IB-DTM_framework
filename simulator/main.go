package main

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/core"
	"time"
)

var ctx, cancel = context.WithCancel(context.Background())

func main() {
	// init all logger at startup
	logutil.InitLogger()

	logutil.LoggerList["main"].Debugf("entering main")
	cfg := config.GenYangNetConfig()
	cfg.ConfigSetGensis(time.Now().Add(3*time.Second))

	session := core.PrepareSimulationSession(cfg)
	go session.Run(ctx)

	time.Sleep(120 * time.Second)
	cancel()
}
