package vehicle

import (
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"math/rand"
)

type Position struct {
	X int
	Y int
}

// defined the data structure in a life cycle
type Vehicle struct {
	// unique id of each vehicle
	Id uint64

	// current Pos of the vehicle
	// set to nil for inactive
	Pos Position
	// the Path of vehicle movement, represented by pos
	// reset when becomes inactive
	// don't know how to maintain it, so just leave it
	// Path []simmap.Position

	// vehicle status
	VehicleStatus int
	// last 6 movement of the vehicle
	LastMovementDirection int
}

// vehiclestatus
const (
	InActive = iota // temporary leave the map, but the data structure remains
	Active          // in the map right now
)

// DEPRECATED! every vehicles' position and movement of previous period should be saved.
//func (v *Vehicle) ResetVehicle() {
//	v.Pos = Position{}
//	v.LastMovementDirection = NotMove
//}

// exceed boundary test is executed by the caller function
// Move helper helps move the vehicle Pos,
// unregister the vehicle from the
func (v *Vehicle) MoveHelper(direction int) {
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
	}
	// update the vehicle's status accordingly
	v.LastMovementDirection = direction
}

// init vehicle, generate the position and lastmovement
// IMPORTANT! the vehicle is not being placed into the map yet!
// IMPORTANT! the syncing with active vehicle bitmap relies on the caller!
func InitVehicle(
	id uint64, // id of the vehicle
	xlen, ylen int, // the size of the map
	active	int,
	r *rand.Rand, // random generator provided by the caller
	) *Vehicle {

	v := &Vehicle{}
	v.Id, v.VehicleStatus = id, active

	// generate the position based on map size
	v.initPosition(r, xlen, ylen)
	return v
}

// this helper function generate position for vehicle and update the vehicle object
func (v *Vehicle) initPosition(r *rand.Rand, xlen, ylen int){
	x, y := randutil.RandIntRange(r, 0, xlen), randutil.RandIntRange(r, 0, ylen)
	v.Pos = Position{x, y}

	denominator, lower, upper := 4, 1, 3
	xLeftBound, xRightBound, yLeftBound, yRightBound :=
		xlen * lower / denominator, xlen * upper / denominator,
		ylen * lower / denominator, ylen * upper / denominator

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
		v.LastMovementDirection = NotMoveGroup[randutil.RandIntRange(r, 0, len(NotMoveGroup))]
	}
}

// return the decision of direction,
// the boundary check logic and offmap logic should be achieved by the caller
func (v *Vehicle) MovementDecisionMaker(r *rand.Rand, xlen, ylen int) int {
	var direction int
	ld := v.LastMovementDirection

	// WARNING! the length of array in direction map is hard coded to 2!
	r1, r2 := r.Float32(), randutil.RandIntRange(r, 0, 2)
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