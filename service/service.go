package service

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/dtm"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator"
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

	// fire up each components
	go func() {
		serviceList["simulation"] = simulator.Run(ctx)
	}()
	go func() {
		serviceList["dtm"] = dtm.Run(ctx)
	}()
	go func() {
		serviceList["statistics"] = statistics.Run(ctx)
	}()
}

func Done() {
	simulator.Done(serviceList["simulation"].(*simulator.SimulationSession))
	statistics.Done()
}
