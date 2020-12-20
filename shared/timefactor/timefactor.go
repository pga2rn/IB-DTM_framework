package timefactor

// TODO: decide the functions, and implement them

import (
	"errors"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"math"
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

var funcNums = 5
var funcTable table

func InitTimeFactor(cfg *config.Config) {
	funcTable = table{}
	length := cfg.SlotsPerEpoch
	funcTable.length = int(length)

	funcTable.funcTable[Exp] = make([]float32, length)
	funcTable.funcTable[Linear] = make([]float32, length)
	funcTable.funcTable[Power] = make([]float32, length)
	funcTable.funcTable[Sin] = make([]float32, length)
	funcTable.funcTable[Log] = make([]float32, length)
}

// print the table
func calculateTimeFactor(cfg config.Config) {
	for i := 0; i < int(cfg.SlotsPerEpoch); i++ {
		x := 0.0
		funcTable.funcTable[Sin][i] = float32(sinFunc(x))
		funcTable.funcTable[Power][i] = float32(powerFunc(x))
		funcTable.funcTable[Log][i] = float32(logFunc(x))
		funcTable.funcTable[Exp][i] = float32(expFunc(x))
		funcTable.funcTable[Linear][i] = float32(x)
	}
}

// y = sin(1/2 * pi * x)
func sinFunc(x float64) float64 {
	return math.Sin(x * 0.5 * math.Pi)
}

// y = 2^x - 1
func expFunc(x float64) float64 {
	return math.Pow(2, x) - 1
}

// y = x^2
func powerFunc(x float64) float64 {
	return math.Pow(x, 2)
}

// y = -0.5log(1/x)+1
func logFunc(x float64) float64 {
	return -1*0.5*math.Log(1/x) + 1
}

func GetTimeFactor(funcType, index int) (float32, error) {
	if index < 0 || index > funcTable.length || funcType < 0 || funcType > funcNums {
		return -1, errors.New("invalid arguments")
	}
	return funcTable.funcTable[funcType][index], nil
}
