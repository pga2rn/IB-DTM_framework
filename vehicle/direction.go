package vehicle

// move direction
const (
	NotMove   = iota
	XForward	// X + 1
	XBackward	// X - 1
	YForward	// Y + 1
	YBackward	// Y - 1
	XFYF		// X + 1, Y + 1
	XFYB		// X + 1, Y - 1
	XBYF		// X - 1, Y + 1
	XBYB		// X - 1, Y - 1
	YFXF	=	XFYF
	YFXB	=	XBYF
	YBXF	= 	XFYB
	YBXB	=	XBYB
)

// possibility const
const (
	KeepStraightDirection = 0.5
	SectorDirection       = 0.4
	LeftOrRightDirection  = 0.1
	AnyDirection          = 1
)

// direction map
var DirectionMap = map[int]*map[float32][]int{
	XForward: &XForwardGroup,
	XBackward: &XBackwardGroup,
	YForward: &YForwardGroup,
	YBackward: &YBackwardGroup,
	XFYF: &XFYFGroup,
	XFYB: &XFYBGroup,
	XBYF: &XBYFGroup,
	XBYB: &XBYBGroup,
}

// supplemented arrays for direction decision
// if lastmovement is not move, every direction is possible
var NotMoveGroup = []int{
	NotMove,
	XForward, XBackward,
	YForward, YBackward,
	XFYF, XFYB,
	XBYF, XBYB,
}

// totally 8 directions
var XForwardGroup = map[float32][]int{
	SectorDirection:{XFYB, XFYF},
	LeftOrRightDirection: {YForward, YBackward},
}

var XBackwardGroup = map[float32][]int{
	SectorDirection: {XBYB, XBYF},
	LeftOrRightDirection: {YForward, YBackward},
}

var YForwardGroup = map[float32][]int{
	SectorDirection: {YFXF, YFXB},
	LeftOrRightDirection: {XForward, XBackward},
}

var YBackwardGroup = map[float32][]int{
	SectorDirection: {YBXB, YBXF},
	LeftOrRightDirection: {XForward, XBackward},
}

// xf
var XFYFGroup = map[float32][]int{
	SectorDirection: {XForward, YForward},
	LeftOrRightDirection: {XFYB, XBYF},
}

var XFYBGroup = map[float32][]int{
	SectorDirection: {XForward, YBackward},
	LeftOrRightDirection: {XFYF, XBYB},
}

// xb
var XBYFGroup = map[float32][]int{
	SectorDirection: {XBackward, YForward},
	LeftOrRightDirection: {XFYF, XBYB},
}

var XBYBGroup = map[float32][]int{
	SectorDirection: {XBackward, YBackward},
	LeftOrRightDirection: {XFYB, XBYF},
}

