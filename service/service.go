package service

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/urfave/cli/v2"
	"reflect"
)

type Services struct {
	ctx          context.Context
	servicesList map[string]interface{}
}

var services Services

// logger init, simconfig init
// TODO: init all services here!
func Init(uCtx *cli.Context) error {
	// init the logger
	logutil.InitLogger()
	logutil.LoggerList["service"].Debugf("init logger")
	// init the package global services object
	services = Services{
		ctx:          uCtx.Context,
		servicesList: make(map[string]interface{}),
	}
	return nil
}

// all the services are initialized here
// channels between each simulator should be assigned here
func Entry(ctx *cli.Context) error {
	// derive context from urfave's cli.Contexts
	// init all logger at startup
	logutil.LoggerList["service"].Debugf("fire up all services")

	// fire up each components
	for name, component := range services.servicesList {
		logutil.LoggerList["service"].Debugf("[run] fire up %v service", name)
		param := []reflect.Value{reflect.ValueOf(ctx.Context)}
		reflect.ValueOf(&component).MethodByName("Run").Call(param)
	}

	// wait for upper caller issuing cancel
	<-ctx.Context.Done()
	return nil
}

func Done(ctx *cli.Context) error {
	//simulator.Done(serviceList["simulation"].(*simulator.SimulationSession))
	//statistics.Done()
	logutil.LoggerList["service"].Debugf("application terminated")

	// fire up each components
	for name, component := range services.servicesList {
		logutil.LoggerList["service"].Debugf("[Done] terminate %v service", name)
		reflect.ValueOf(&component).MethodByName("Done").Call([]reflect.Value{})
	}

	// wait for upper caller issuing cancel
	<-ctx.Context.Done()
	return nil
}
