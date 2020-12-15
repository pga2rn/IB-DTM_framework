package main

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/core"
	"github.com/sirupsen/logrus"
	"time"
)

var ctx, cancel = context.WithCancel(context.Background())
var log = logrus.WithField("prefix", "main")

func main() {
	log.Info("Main process starts..")
	cfg := config.GenYangNetConfig()
	cfg.ConfigSetGensis(time.Now().Add(3*time.Second))

	session := core.PrepareSimulationSession(cfg)
	go session.Run(ctx)

	time.Sleep(20 * time.Second)
	cancel()
}
