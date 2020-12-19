package timefactor

// TODO: decide the functions, and implement them

import (
	"errors"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
)

// pre-calculated values
var ExponentialFunction *[]int // y = 2^x - 1
var LinearFunction *[]int // y = x
var PowerFunction *[]int // y = x^n
var SinFunction *[]int // y = sin(1/2 * pi * x)
var LogarithmFunction *[]int // y = -nlog(1/x)+1, smaller the n, the steeper when x reaches 0

func InitTimeFactor(cfg config.Config){
	slotsPerEpoch := cfg.SlotsPerEpoch

}

func GetExp(index int) (int, error){
	if ExponentialFunction == nil || index > len(*ExponentialFunction) || index < 0 {
		return -1, errors.New("invalid arguments")
	}
	return (*ExponentialFunction)[index], nil
}

func GetLinear(index int) (int, error){
	if LinearFunction == nil || index < 0 {
		return -1, errors.New("invalid arguments")
	}
	return index, nil
}

func GetPower(index int) (int, error){
	if PowerFunction == nil || index > len(*PowerFunction) || index < 0 {
		return -1, errors.New("invalid arguments")
	}
	return (*PowerFunction)[index], nil
}