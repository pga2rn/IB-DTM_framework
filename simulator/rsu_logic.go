package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/pair"
	"sync"
)

func (sim *SimulationSession) InitRSUs() bool {
	for x := range sim.RSUs {
		// init every RSU data structure
		for y := range sim.RSUs[x] {
			r := rsu.InitRSU(
				uint32(sim.CoordToIndex(x, y)),
				pair.Position{x, y},
				sim.Config.RingLength,
			)

			// uploading tracker
			r.SetNextUploadSlot(0)
			sim.RSUs[x][y] = r
		}
	}

	return true
}

// helper function for processEpoch to assign compromisedRSU
func (sim *SimulationSession) initAssignCompromisedRSU(ctx context.Context) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[initAssignMisbehaveVehicle] context canceled")
		return
	default:
		sim.CompromisedRSUPortion = sim.Config.CompromisedRSUPortion

		target := int(float32(sim.Config.RSUNum) * sim.CompromisedRSUPortion)

		for i := 0; i < target; i++ {
			index := sim.R.RandIntRange(0, sim.Config.RSUNum)
			if !sim.CompromisedRSUBitMap.Get(index) {
				sim.CompromisedRSUBitMap.Set(index, true)
			}
		}
	}
}

// Evil type 1: alter the existed trust value offsets
// the altered trust value offsets will finally being altered when trust values are being calculated
// dive into the slot
func (sim *SimulationSession) alterTrustValueOffset(ctx context.Context, rsu *rsu.RSU, slot uint32) {
	tvoStorage := rsu.GetSlotInRing(slot)

	c := make(chan []interface{})
	// define a call back function to take the value out of sync.map
	f := func(key, value interface{}) bool {
		c <- []interface{}{key, value}
		return true
	}

	wg := sync.WaitGroup{}

	go func() {
		wg.Add(1)
		select {
		case <-ctx.Done():
			logutil.LoggerList["simulator"].Fatalf("[alterTrustValueOffset] go routine context canceled, abort")
		default:
			// iterate through the slot storage of RSU
			for value := range c {
				_, tvo := value[0].(uint32), value[1].(*dtmtype.TrustValueOffset)

				// if the RSU is compromised, decide which type of evil it will do to the tvo
				rn := sim.R.Float32()
				// assign altered type
				if tvo.TrustValueOffset < 0 {
					if rn < 0.8 {
						tvo.AlterType = dtmtype.Flipped
					} else {
						tvo.AlterType = dtmtype.Dropped
					}
				} else {
					tvo.AlterType = dtmtype.Flipped
				}
			}
		}
		wg.Done()
	}()

	tvoStorage.Range(f)
	close(c)
	wg.Wait()
}

// evil type 2
// Evil type 2, forge trust value offsets
// store the updated tvo back to RSU
// RSU will try to make more vehicles being treated as misbehaving,
// to over thrown the dtm itself
func (sim *SimulationSession) forgeTrustValueOffset(ctx context.Context, rsu *rsu.RSU, slot uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[forgeTrustValueOffset] rsu managed v %v, context canceled", rsu.ManagedVehicles)
	default:
		rn, target := sim.R.Float32(), 0
		// if managed vehicles num is too small
		// the compromised RSU will not do evils to hide themselves
		if rsu.ManagedVehicles < sim.ActiveVehiclesNum/sim.Config.RSUNum {
			return
		} else {
			target = rsu.ManagedVehicles
		}

		switch {
		case rn < 0.8:
			target = target / 5
		case rn < 0.3:
			target = target * 2 / 5
		default:
			target = sim.R.RandIntRange(target, target*3/2)
		}

		for i := 0; i < target; {
			vid := uint32(sim.R.RandIntRange(0, sim.Config.VehicleNumMax))

			if sim.Map.GetCross(rsu.Pos).CheckIfVehicleInManagementZone(vid) {
				continue
			} else {
				i++
			}

			tvo := &dtmtype.TrustValueOffset{
				AlterType: dtmtype.Forged,
				VehicleId: vid,
			}

			// randomly rate the vehicle
			rn := sim.R.Float32()
			switch {
			case rn < 0.5:
				tvo.TrustValueOffset = -1
			default:
				tvo.TrustValueOffset = 1
			}

			rn = sim.R.Float32()
			switch {
			case rn < 0.2:
				tvo.Weight = dtmtype.Fatal
			case rn < 0.4:
				tvo.Weight = dtmtype.Critical
			default:
				tvo.Weight = dtmtype.Routine
			}

			// store the forged data into RSU storage area
			rsu.GetSlotInRing(slot).Store(vid, tvo)
		}
	}
}
