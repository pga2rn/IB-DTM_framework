package service

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/core"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/statistics"
)

var serviceList = make(map[string]interface{})

func Init() {
	// init the logger
	logutil.InitLogger()
	logutil.LoggerList["service"].Debugf("init logger and services")
}

func Run(ctx context.Context) {
	// init all logger at startup
	logutil.LoggerList["service"].Debugf("fire up all services")

	// fire up simulation session
	go func() {
		serviceList["simulation"] = core.Run(ctx)
	}()
	go func() {
		serviceList["statistics"] = statistics.Run(ctx)
	}()
}

func Done() {
	core.Done(serviceList["simulation"].(*core.SimulationSession))
	statistics.Done()
}
