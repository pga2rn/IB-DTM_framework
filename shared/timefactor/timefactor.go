package timefactor

// TODO: decide the functions, and implement them

import (
	"math"
	"time"
)

// pre-calculated values
type table struct {
	length    int
	funcTable map[int][]float32
}

const (
	Exp = iota
	Linear
	Power
	Sin
	Log
)

func GetTimeFactor(timeFactorType int, genesis time.Time, slotTime time.Time, checkPointTime time.Time) float64 {
	x := float64(checkPointTime.Unix()-genesis.Unix()) / float64(slotTime.Unix()-genesis.Unix())
	res := float64(-1)
	switch timeFactorType {
	case Exp: // y = 2^x - 1
		res = math.Pow(2, x) - 1
	case Linear: // y = x
		res = x
	case Power: // y = x^2
		res = math.Pow(x, 2)
	case Sin: // y = sin(1/2 * pi * x)
		res = math.Sin(0.5 * math.Pi * x)
	case Log: // y = -0.5log(1/x)+1
		res = -1*0.5*math.Log(1/x) + 1
	}
	return res
}
