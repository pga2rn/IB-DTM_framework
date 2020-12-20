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
	exp = iota
	linear
	power
	sin
	log
)

var funcNums = 5
var funcTable table

func InitTimeFactor(cfg *config.Config) {
	funcTable = table{}
	length := cfg.SlotsPerEpoch
	funcTable.length = int(length)

	funcTable.funcTable[exp] = make([]float32, length)
	funcTable.funcTable[linear] = make([]float32, length)
	funcTable.funcTable[power] = make([]float32, length)
	funcTable.funcTable[sin] = make([]float32, length)
	funcTable.funcTable[log] = make([]float32, length)
}

// print the table
func calculateTimeFactor(cfg config.Config) {
	for i := 0; i < int(cfg.SlotsPerEpoch); i++ {
		x := 0.0
		funcTable.funcTable[sin][i] = float32(sinFunc(x))
		funcTable.funcTable[power][i] = float32(powerFunc(x))
		funcTable.funcTable[log][i] = float32(logFunc(x))
		funcTable.funcTable[exp][i] = float32(expFunc(x))
		funcTable.funcTable[linear][i] = float32(x)
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
