package service

import (
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/dtm"
	ib_dtm "github.com/pga2rn/ib-dtm_framework/ib-dtm"
	"github.com/pga2rn/ib-dtm_framework/rpc"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"reflect"
	"time"
)

type Services map[string]interface{}

var PackageName = "service"

var services = make(Services)

// logger init, simconfig init
func Init(uCtx *cli.Context) error {

	// init the logger
	logutil.SetLevel(logrus.DebugLevel)
	logutil.GetLogger(PackageName).Debugf("[Init] init logger")

	// init the simulation config
	config.SetGenesis(time.Now().Add(3 * time.Second))
	logutil.GetLogger(PackageName).Debugf("[Init] genesis will kick after %v seconds", 6)

	//// init the channel for intercommunication
	sim2DTM := make(chan interface{})
	sim2IBDTM := make(chan interface{})
	DTM2IBDTM := make(chan interface{})
	DTM2RPC := make(chan interface{})

	// init and register the services
	services["simulator"] = simulator.PrepareSimulationSession(
		config.GenYangNetConfig(),
		sim2DTM, sim2IBDTM)
	services["dtm"] = dtm.PrepareDTMLogicModuleSession(
		config.GenYangNetConfig(),
		config.InitExperimentConfig(),
		sim2DTM, DTM2IBDTM, DTM2RPC)
	services["ib-dtm"] = ib_dtm.PrepareBlockchainModule(
		config.GenYangNetConfig(),
		config.InitProposalExperimentConfigList(),
		config.GenIBDTMConfig(config.GenYangNetConfig()),
		sim2IBDTM, DTM2IBDTM)
	services["rpc"] = rpc.PrepareRPCServer(DTM2RPC)
	//services["statistics"] = statistics.PrepareStatisticsSession(statisticsCfg, expCfg)

	logutil.GetLogger(PackageName).Debugf("[Init] finished registering services")
	logutil.SetServiceList(services)
	return nil
}

// all the services are initialized here
// channels between each simulator should be assigned here
func Entry(ctx *cli.Context) error {
	// derive context from urfave's cli.Contexts
	// init all logger at startup
	logutil.GetLogger(PackageName).Debugf("[Entry] main routine starts")

	// fire up each modules via Run function
	for name, component := range services {
		logutil.GetLogger(PackageName).Debugf("[Entry] fire up %v service", name)
		param := []reflect.Value{reflect.ValueOf(ctx.Context)}
		go reflect.ValueOf(component).MethodByName("Run").Call(param)
	}

	// wait for upper caller issuing cancel
	<-ctx.Done()
	return nil
}

func Done(ctx *cli.Context) error {
	logutil.GetLogger(PackageName).Debugf("application terminated")

	// call each module's termination functions
	for name, component := range services {
		logutil.GetLogger(PackageName).Debugf("[Done] terminate %v service", name)
		param := []reflect.Value{reflect.ValueOf(ctx.Context)}
		go reflect.ValueOf(&component).MethodByName("Done").Call(param)
	}

	// wait for upper caller issuing cancel
	<-ctx.Done()
	return nil
}
