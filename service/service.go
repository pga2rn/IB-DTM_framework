package service

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/dtm"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator"
	"github.com/urfave/cli/v2"
	"reflect"
	"time"
)

type Services struct {
	ctx          context.Context
	servicesList map[string]interface{}
}

var services Services

// logger init, simconfig init
func Init(uCtx *cli.Context) error {
	// init the logger
	logutil.InitLogger()
	logutil.LoggerList["service"].Debugf("[Init] init logger")
	// init the package global services object
	services = Services{
		ctx:          uCtx.Context,
		servicesList: make(map[string]interface{}),
	}

	// init the simulation config
	cfg := config.GenYangNetConfig()
	cfg.SetGenesis(time.Now().Add(2 * time.Second))
	logutil.LoggerList["service"].Debugf("[Init] genesis will kick after 2 seconds")

	// init experiment config
	expCfg := config.InitBaselineExperimentConfig()

	// init the channel for intercommunication
	simDTMcomm := make(chan interface{})

	// init and register the services
	services.servicesList["simulator"] = simulator.PrepareSimulationSession(cfg, simDTMcomm)
	services.servicesList["dtm"] = dtm.PrepareDTMLogicModuleSession(cfg, expCfg, simDTMcomm)

	logutil.LoggerList["service"].Debugf("[Init] register of services finished")
	return nil
}

// all the services are initialized here
// channels between each simulator should be assigned here
func Entry(ctx *cli.Context) error {
	// derive context from urfave's cli.Contexts
	// init all logger at startup
	logutil.LoggerList["service"].Debugf("[Entry] main routine starts")

	// fire up each components
	for name, component := range services.servicesList {
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

	// fire up each components
	for name, component := range services.servicesList {
		logutil.LoggerList["service"].Debugf("[Done] terminate %v service", name)
		go reflect.ValueOf(&component).MethodByName("Done").Call([]reflect.Value{})
	}

	// wait for upper caller issuing cancel
	<-ctx.Done()
	return nil
}
