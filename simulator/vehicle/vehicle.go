package vehicle

// TODO: error handle and context

type Position struct {
	X int
	Y int
}

// defined the data structure in a life cycle
type Vehicle struct {
	// unique id of each vehicle
	id uint64

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
	InActive  = iota  // temporary leave the map, but the data structure remains
	Active      // in the map right now
)

// move direction
const (
	NotMove = 0
	XForward = 1
	XBackward = -XForward
	YForward = 2
	YBackward = -YForward
)
var DirectionArray = []int{XForward, XBackward, YForward, YBackward}