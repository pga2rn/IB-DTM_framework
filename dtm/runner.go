package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

func Run(ctx context.Context) {
	session := DTMLogicSession{}
	session.run(ctx)
}

func (session *DTMLogicSession) done() {
	close(session.ChanBlockchain)
	close(session.ChanSim)
}

// init the experiments
// each experiment setup will be initialized
//func InitDTMSession(
//	simConfig config.SimConfig, experimentConfig []*config.ExperimentConfig,
//	vehicles *[]*vehicle.Vehicle, vmu *sync.Mutex,
//	rsus *[][]*RSU, rmu *sync.Mutex,
//	mVehicleBitMap *bitmap.Threadsafe,
//	) *map[string]*DTMLogicSession {
//
//	sessions := make(map[string]*DTMLogicSession)
//
//	for _, e := range experimentConfig{
//		session := &DTMLogicSession{}
//		session.Config = e
//		session.SimConfig = &simConfig
//
//		// get the pointer and mutex
//		// TODO: I don't know it is a good idea to do so!
//		session.Vehicles = vehicles
//		session.vmu = vmu
//		session.RSUs = rsus
//		session.rmu = rmu
//
//		// get the correct answer
//		session.MisbehavingVehicleBitMap = mVehicleBitMap
//
//		// init the trust value storage
//		session.TrustValueStorageHead = dtmtype.InitTrustValueStorage()
//
//		sessions[e.Name] = session
//	}
//
//	return &sessions
//}

func (session *DTMLogicSession) run(ctx context.Context) {
	// wait for simulator to call & initialized the dtm logic module
	//if err := session.WaitForSimulator(); err != nil {
	//	session.done()
	//	logutil.LoggerList["dtm"].Fatalf("failed to wait for simulator start")
	//}

	// after initialization is finished, waiting for the communication from the simulator
	for {
		select {
		case <-ctx.Done():
			session.done()
			return
		case v := <-session.ChanSim:
			// unpack
			pack := v.(shared.SimDTMCommunication)
			session.Slot, session.Epoch = pack.Slot, pack.Slot/session.SimConfig.SlotsPerEpoch
			session.CompromisedRSUBitMap = pack.CompromisedRSUBitMap

			logutil.LoggerList["dtm"].Debugf("[run] slot %v, epoch %v", session.Slot, session.Epoch)
			// TODO: keep finishing the dtm logic
		default:
			logutil.LoggerList["dtm"].Debugf("[run] chansim closed, abort")
			session.done()
			return
		}
	}
}
