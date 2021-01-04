package rpc

import (
	"context"
	"flag"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	gw "github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

type RPCServerSession struct {
	serverLis  string
	gatewayLis string
	chanDTM    chan interface{}

	serverInstance  *grpc.Server
	gwInstance      *http.Server
	cacheLatestData *pb.StatisticsBundle
}

var ServerSession *RPCServerSession

func PrepareRPCServer(dtm chan interface{}) *RPCServerSession {
	session := &RPCServerSession{
		serverLis:       "localhost:5000",
		gatewayLis:      "0.0.0.0:5001",
		chanDTM:         dtm,
		cacheLatestData: &pb.StatisticsBundle{},
	}
	ServerSession = session
	return session
}

func (rpcs *RPCServerSession) GetLatestData() *pb.StatisticsBundle {
	return rpcs.cacheLatestData
}

func (rpcs *RPCServerSession) startRPCServer(ctx context.Context) {

	lisparam := rpcs.serverLis
	lis, err := net.Listen("tcp", lisparam)
	if err != nil {
		logutil.LoggerList["rpc"].Fatal("[Run] failed to init rpc server")
	}

	s := grpc.NewServer()
	rpcs.serverInstance = s

	pb.RegisterFrameworkStatisticsQueryServer(s, &Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			logutil.LoggerList["rpc"].Fatal("[Run] failed to start the server")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// if the context is canceled, stop the server and exit the go routine
			s.Stop()
			return
		}
	}
}

func (rpcs *RPCServerSession) startRPCgw(ctx context.Context) {
	// create server backend endpoint
	grpcServerEndpoint := flag.String("grpc-server-endpoint", rpcs.serverLis, "gRPC server endpoint")

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	// allow cors
	mux := runtime.NewServeMux()
	newMux := handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
	)(mux)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterFrameworkStatisticsQueryHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		logutil.LoggerList["rpc"].Fatalf("[startRPCgw] error when register the handler to gateway server")
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	gws := &http.Server{
		Handler: newMux,
		Addr:    rpcs.gatewayLis,
	}
	rpcs.gwInstance = gws
	if err := gws.ListenAndServe(); err != nil {
		logutil.LoggerList["rpc"].Fatalf("[startRPCgw] error when start the gateway server")
	}
}

func (rpcs *RPCServerSession) Done(ctx context.Context) {
	logutil.LoggerList["rpc"].Debugf("[Done] terminate the RPC server")
	rpcs.serverInstance.GracefulStop()
	if err := rpcs.gwInstance.Shutdown(ctx); err != nil {
		return
	}
}

func (rpcs *RPCServerSession) Run(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		go rpcs.startRPCServer(ctx)
		go rpcs.startRPCgw(ctx)

		logutil.LoggerList["rpc"].Infof("[Run] framework query server now listens at %v", rpcs.gatewayLis)
		logutil.LoggerList["rpc"].Infof("[Run] API: http://127.0.0.1:5001/v1/framework/data")

		// start the main routine loop
		// listen the chandtm channel and waiting for latest available data
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-rpcs.chanDTM:
				logutil.LoggerList["rpc"].Debugf("[Run] receive data from dtm")
				rpcs.cacheLatestData = data.(*pb.StatisticsBundle)
			}
		}
	}
}
