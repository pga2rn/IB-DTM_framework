package dtm

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// communicating with simulator: RSU compromised bitmap, slot

type DTMLogicSession struct {
	// configs
	ExpConfig *map[string]*config.ExperimentConfig
	SimConfig *config.SimConfig
	// the correct answer
	MisbehavingVehicleBitMap *bitmap.Threadsafe
	CompromisedRSUBitMap     *bitmap.Threadsafe

	// session status
	Slot, Epoch uint32

	// pointer to the vehicles and RSU
	// I don't know if it is a good idea to use mutex via pointer
	Vehicles *[]*vehicle.Vehicle
	RSUs     *[][]*rsu.RSU
	vmu      *sync.Mutex
	rmu      *sync.Mutex

	// channel
	ChanSim        chan interface{}
	ChanRPC        chan interface{}
	ChanBlockchain chan interface{}

	// trust value storage and misbehaving flag results for epochs
	// each experiment instance has its own trust value storage
	TrustValueStorageHead *map[string]*dtmtype.TrustValueStorageHead

	// storage area for proposal
	ProposalStorage *map[string]*IBDTMStorage

	// a random source
	R *randutil.RandUtil
}

func (session *DTMLogicSession) informRPCServer(ctx context.Context, epoch uint32) {
	select {
	case <-ctx.Done():
		return
	default:
		expNum := len(*session.ExpConfig)
		statisticsBundle := &pb.StatisticsBundle{
			Epoch: epoch, Bundle: make([]*pb.StatisticsPerExperiment, expNum),
		}

		// query the newest results
		count := 0
		for expName, _ := range *session.ExpConfig {
			head := (*session.TrustValueStorageHead)[expName]
			if ep, _ := head.GetEpochInformation(); ep != epoch {
				logutil.LoggerList["dtm"].Debugf("[informRPCServer] epoch async")
				continue
			}

			// add results to statistics bundle
			statisticsBundle.Bundle[count] = head.GetHeadBlock().GetStatistics()
			count++
		}

		// pass the bundle to the rpc server
		select {
		case session.ChanRPC <- statisticsBundle:
		default:
			logutil.LoggerList["dtm"].Debugf("[informRPCServer] transimition dropped")
		}
	}
}

func PrepareDTMLogicModuleSession(
	simCfg *config.SimConfig, expCfg *map[string]*config.ExperimentConfig,
	simdtmChan chan interface{}, dtmrpcChan chan interface{},
) *DTMLogicSession {

	dtmSession := &DTMLogicSession{}

	// init the simulator config and experiment config
	dtmSession.SimConfig = simCfg
	dtmSession.ExpConfig = expCfg

	// inter module communication
	dtmSession.ChanSim = simdtmChan
	dtmSession.ChanRPC = dtmrpcChan

	// random source
	dtmSession.R = randutil.InitRand(123)

	// prepare experiments
	dtmSession.TrustValueStorageHead = dtmtype.InitTrustValueStorageHeadMap()
	for expName, exp := range *expCfg {
		(*dtmSession.TrustValueStorageHead)[expName] = dtmtype.InitTrustValueStorage()
		if exp.Type == pb.ExperimentType_PROPOSAL {
			dtmSession.ProposalStorage = InitIBDTMStorageMap()
		}
	}

	return dtmSession
}
