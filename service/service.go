package service

import (
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/dtm"
	ib_dtm "github.com/pga2rn/ib-dtm_framework/ib-dtm"
	"github.com/pga2rn/ib-dtm_framework/rpc"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator"
	"github.com/urfave/cli/v2"
	"reflect"
	"time"
)

type Services map[string]interface{}

var services = make(Services)

// logger init, simconfig init
func Init(uCtx *cli.Context) error {
	cfg := config.GenYangNetConfig()

	// init the logger
	logutil.InitLogger(cfg.Loglevel)
	logutil.LoggerList["service"].Debugf("[Init] init logger")

	// init the simulation config
	cfg.SetGenesis(time.Now().Add(time.Duration(cfg.SecondsPerSlot) * time.Second))
	logutil.LoggerList["service"].Debugf("[Init] genesis will kick after %v seconds", cfg.SecondsPerSlot)

	// init experiment config
	expCfg := config.InitExperimentConfig()
	//
	//// init the channel for intercommunication
	simDTMComm := make(chan interface{})
	simBCComm := make(chan interface{})
	DTMBCComm := make(chan interface{})
	DTMRPCComm := make(chan interface{})

	// init and register the services
	services["simulator"] = simulator.PrepareSimulationSession(cfg, simDTMComm)
	services["dtm"] = dtm.PrepareDTMLogicModuleSession(cfg, expCfg, simDTMComm, DTMRPCComm)
	services["ib-dtm"] = ib_dtm.PrepareBlockchainModule(cfg, simBCComm, DTMBCComm)
	services["rpc"] = rpc.PrepareRPCServer(DTMRPCComm)

	logutil.LoggerList["service"].Debugf("[Init] finished registering services")
	return nil
}

// all the services are initialized here
// channels between each simulator should be assigned here
func Entry(ctx *cli.Context) error {
	// derive context from urfave's cli.Contexts
	// init all logger at startup
	logutil.LoggerList["service"].Debugf("[Entry] main routine starts")

	// fire up each modules via Run function
	for name, component := range services {
		logutil.LoggerList["service"].Debugf("[Entry] fire up %v service", name)
		param := []reflect.Value{reflect.ValueOf(ctx.Context)}
		go reflect.ValueOf(component).MethodByName("Run").Call(param)
	}

	// wait for upper caller issuing cancel
	<-ctx.Done()
	return nil
}

func Done(ctx *cli.Context) error {
	logutil.LoggerList["service"].Debugf("application terminated")

	// call each module's termination functions
	for name, component := range services {
		logutil.LoggerList["service"].Debugf("[Done] terminate %v service", name)
		param := []reflect.Value{reflect.ValueOf(ctx.Context)}
		go reflect.ValueOf(&component).MethodByName("Done").Call(param)
	}

	// wait for upper caller issuing cancel
	<-ctx.Done()
	return nil
}
