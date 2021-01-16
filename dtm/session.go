package dtm

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"sync"
)

var PackageName = "dtm"

// communicating with simulator: RSU compromised bitmap, slot

type DTMLogicSession struct {
	// configs
	ExpConfig map[string]*config.ExperimentConfig
	SimConfig *config.SimConfig
	// the correct answer
	MisbehavingVehicleBitMap *bitmap.Threadsafe
	CompromisedRSUBitMap     *bitmap.Threadsafe

	// session status
	Slot, Epoch       uint32
	ActiveVehiclesNum int32

	// pointer to the vehicles and RSU
	RSUs [][]*rsu.RSU
	rmu  *sync.Mutex

	// channel
	ChanSim   chan interface{}
	ChanRPC   chan interface{}
	ChanIBDTM chan interface{}

	// trust value storage and misbehaving flag results for epochs
	// each experiment instance has its own trust value storage
	TrustValueStorageHead map[string]*fwtype.TrustValueStorageHead

	// a random source
	R *randutil.RandUtil
}

func (session *DTMLogicSession) informRPCServer(ctx context.Context, epoch uint32) {
	select {
	case <-ctx.Done():
		return
	default:
		expNum := len(session.ExpConfig)
		statisticsBundle := &pb.StatisticsBundle{
			Epoch:             epoch,
			Bundle:            make([]*pb.StatisticsPerExperiment, expNum),
			ActiveVehicleNums: session.ActiveVehiclesNum,
		}

		// query the newest results
		count := 0
		for expName, _ := range session.ExpConfig {
			head := session.TrustValueStorageHead[expName]
			if ep, _ := head.GetEpochInformation(); ep != epoch {
				logutil.GetLogger(PackageName).Debugf("[informRPCServer] epoch async")
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
			logutil.GetLogger(PackageName).Debugf("[informRPCServer] transimition dropped")
		}
	}
}

func PrepareDTMLogicModuleSession(
	simCfg *config.SimConfig, expCfg map[string]*config.ExperimentConfig,
	simChan, ibdtmChan, rpcChan chan interface{},
) *DTMLogicSession {
	dtmSession := &DTMLogicSession{
		// init the simulator config and experiment config
		SimConfig: simCfg,
		ExpConfig: expCfg,

		// inter module communication
		ChanSim:   simChan,
		ChanRPC:   rpcChan,
		ChanIBDTM: ibdtmChan,

		// random source
		R: randutil.InitRand(123),
	}

	// prepare experiments
	dtmSession.TrustValueStorageHead = make(map[string]*fwtype.TrustValueStorageHead)
	for expName, _ := range expCfg {
		dtmSession.TrustValueStorageHead[expName] = fwtype.InitTrustValueStorage()
	}

	return dtmSession
}
