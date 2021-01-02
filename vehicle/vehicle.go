package vehicle

import (
	"github.com/pga2rn/ib-dtm_framework/shared/pair"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
)

// defined the data structure in a life cycle
type Vehicle struct {
	// unique id of each vehicle
	Id uint32

	// current Pos of the vehicle
	// set to nil for inactive
	Pos pair.Position
	// the Path of vehicle movement, represented by pos
	// reset when becomes inactive
	// don't know how to maintain it, so just leave it
	// Path []simmap.Position

	// vehicle status
	VehicleStatus VehicleStatus
	// last 6 movement of the vehicle
	LastMovementDirection Direction
}

type Direction = int
type VehicleStatus = int

// vehiclestatus
const (
	InActive = iota // temporary leave the map, but the data structure remains
	Active          // in the map right now
)

// init vehicle, generate the position and lastmovement
// IMPORTANT! the vehicle is not being placed into the map yet!
// IMPORTANT! the syncing with active vehicle bitmap relies on the caller!
func InitVehicle(
	id uint32, // id of the vehicle
	xlen, ylen int, // the size of the map
	active VehicleStatus,
	r *randutil.RandUtil, // random generator provided by the caller
) *Vehicle {

	v := &Vehicle{}
	v.Id, v.VehicleStatus = id, active

	// generate the position based on map size
	v.InitPosition(r, xlen, ylen)
	return v
}

func (v *Vehicle) ResetVehicle() {
	v.Pos = pair.Position{}
	v.LastMovementDirection = NotMove
}

// exceed boundary test is executed by the caller function
// Move helper helps move the vehicle Pos,
// unregister the vehicle from the
func (v *Vehicle) VehicleMove(direction Direction) {
	// update pos
	switch direction {
	case XForward:
		v.Pos.X += 1
	case XBackward:
		v.Pos.X -= 1
	case YForward:
		v.Pos.Y += 1
	case YBackward:
		v.Pos.Y -= 1
	case XFYF:
		v.Pos.X += 1
		v.Pos.Y += 1
	case XFYB:
		v.Pos.X += 1
		v.Pos.Y -= 1
	case XBYF:
		v.Pos.X -= 1
		v.Pos.Y += 1
	case XBYB:
		v.Pos.X -= 1
		v.Pos.Y -= 1
	}
	// update the vehicle's status accordingly
	v.LastMovementDirection = direction
}

// this helper function generate the position when the vehicle comes back into the map
// update the vehicle's position and lastmovementdirection
func (v *Vehicle) EnterMap(r *randutil.RandUtil, xlen, ylen int) {
	// x01y0 represents the edge (0, 0)~(1, 0)
	// x0y01 represents the edge (0, 0)~(0, 1)
	// x1y01 represents the edge (1, 0)~(1, 1)
	// y1x01 represents the edge (0, 1)~(1, 1)
	const (
		x01y0 = iota
		y01x0
		x1y01
		y1x01
	)

	edge, pos, direction := r.RandIntRange(0, 4), pair.Position{}, NotMove
	switch edge {
	case x01y0:
		pos.X, pos.Y = r.RandIntRange(0, xlen), 0
		direction = YForward
	case y01x0:
		pos.X, pos.Y = 0, r.RandIntRange(0, ylen)
		direction = XForward
	case x1y01:
		pos.X, pos.Y = 1, r.RandIntRange(0, ylen)
		direction = XBackward
	case y1x01:
		pos.X, pos.Y = r.RandIntRange(0, xlen), 1
		direction = YBackward
	}
	v.Pos, v.LastMovementDirection = pos, direction
}

// this helper function generate position for vehicle and update the vehicle object
func (v *Vehicle) InitPosition(r *randutil.RandUtil, xlen, ylen int) {
	x, y := r.RandIntRange(0, xlen), r.RandIntRange(0, ylen)
	v.Pos = pair.Position{x, y}

	denominator, lower, upper := 4, 1, 3
	xLeftBound, xRightBound, yLeftBound, yRightBound :=
		xlen*lower/denominator, xlen*upper/denominator,
		ylen*lower/denominator, ylen*upper/denominator

	// generate the lastmovement based on the position
	// the map is roughly divided into 5 pieces
	// the idea is to try to let vehicle moving toward the center of the map
	switch {
	case x < xLeftBound && y < yLeftBound:
		v.LastMovementDirection = XFYF
	case x > xRightBound && y > yRightBound:
		v.LastMovementDirection = XBYB
	case x > xRightBound && y < yRightBound:
		v.LastMovementDirection = XBYF
	case x < xLeftBound && y > yRightBound:
		v.LastMovementDirection = XFYB
	case (x > xLeftBound && x < xRightBound) && y < yLeftBound:
		v.LastMovementDirection = YForward
	case (x > xLeftBound && x < xRightBound) && y > yRightBound:
		v.LastMovementDirection = YBackward
	case (y > yLeftBound && y < yRightBound) && x < xLeftBound:
		v.LastMovementDirection = XForward
	case (y > yLeftBound && y < yRightBound) && x > xRightBound:
		v.LastMovementDirection = XBackward
	default:
		// if the vehicle is placed in the center of map
		v.LastMovementDirection = NotMoveGroup[r.RandIntRange(0, len(NotMoveGroup))]
	}
}

// return the decision of direction,
// the boundary check logic and offmap logic should be achieved by the caller
func (v *Vehicle) MovementDecisionMaker(r *randutil.RandUtil, xlen, ylen int) int {
	var direction int
	ld := v.LastMovementDirection

	if ld == NotMove {
		return NotMoveGroup[r.RandIntRange(0, len(NotMoveGroup))]
	}

	// WARNING! the length of array in direction map is hard coded to 2!
	r1, r2 := r.Float32(), r.RandIntRange(0, 2)
	dmap := *DirectionMap[ld]
	switch {
	case r1 > KeepStraightDirection:
		direction = ld
	case r1 <= KeepStraightDirection && r1 > SectorDirection:
		direction = dmap[SectorDirection][r2]
	case r1 <= SectorDirection && r1 > LeftOrRightDirection:
		direction = dmap[LeftOrRightDirection][r2]
	default:
		direction = dmap[SectorDirection][r2]
	}

	return direction
}
