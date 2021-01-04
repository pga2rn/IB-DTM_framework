package rpc

import (
	"context"
	"errors"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedFrameworkStatisticsQueryServer
}

func (s *Server) GetLatestData(ctx context.Context, in *emptypb.Empty) (*pb.StatisticsBundle, error) {
	logutil.LoggerList["rpc"].Debugf("[GetLatestData] received request")
	select {
	case <-ctx.Done():
		return nil, errors.New("context canceled")
	default:
		res := ServerSession.GetLatestData()
		return res, nil
	}
}

func (s *Server) GetDataForEpoch(context.Context, *pb.QueryEpoch) (*pb.StatisticsBundle, error) {
	res := new(pb.StatisticsBundle)
	return res, nil
}

func (s *Server) EchoEpoch(ctx context.Context, in *pb.QueryEpoch) (*pb.QueryEpoch, error) {
	return &pb.QueryEpoch{Epoch: in.Epoch}, nil
}
