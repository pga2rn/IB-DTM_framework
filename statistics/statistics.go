package statistics

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

type Statistics struct {
	Epoch uint64
	// total vehicles num
	VehiclesNum int
	// 4 basic metrics
	TP, FP, FN, TN float64
	// 4 advanced metrics
	Recall, Precision, F1ssimulator, ACC float64
}

// compare the results and calculate the statistics
func GenStatisicsForEpoch(epoch uint64, answer, result *bitmap.Threadsafe) *Statistics {
	length := answer.Len()
	res := &Statistics{Epoch: epoch, VehiclesNum: length}

	// 4 basic metrics
	tp, fp, fn, tn := 0.0, 0.0, 0.0, 0.0
	for i := 0; i < length; i++ {
		a, r := answer.Get(i), result.Get(i)
		switch {
		// result correctly flags out the misbehaving vehicle
		case a == r && a == true:
			tp++
		// result correctly flags out the normal vehicle
		case a == r && a == false:
			tn++
		// result incorrectly flags out the misbehaving vehicle as normal
		case a != r && a == true:
			fn++
		// result incorrectly flags out the normal vehicle as misbehaving
		case a != r && a == false:
			fp++
		}
	}
	res.TP, res.TN, res.FN, res.FP = tp, tn, fn, fp

	// advanced metrics
	recall := tp / (tp + fn)
	precision := tp / (tp + fp)
	f1ssimulator := 2 * recall * precision / (recall + precision)
	acc := (tp + tn) / (tp + tn + fp + fn)
	res.Recall, res.Precision, res.F1ssimulator, res.ACC = recall, precision, f1ssimulator, acc

	return res
}

func Run(ctx context.Context) int {
	logutil.LoggerList["statistics"].Debugf("[Run] start statistics service")
	return 0
}

func Done() int {
	return 0
}
