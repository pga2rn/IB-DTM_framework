package statistics

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

// compare the results and calculate the statistics
func GenStatisticsForEpoch(epoch uint32, answer, result *bitmap.Threadsafe) *pb.StatisticsPerExperiment {
	length := answer.Len()
	res := &pb.StatisticsPerExperiment{Epoch: epoch}

	// 4 basic metrics
	tp, fp, fn, tn := float32(0.0), float32(0.0), float32(0.0), float32(0.0)
	for i := 0; i < length; i++ {
		a, r := answer.Get(i), result.Get(i)
		switch {
		// positive is misbehaving, negative is normal
		// correctly flag misbehaving vehicles
		case a == r && a == true:
			tp++
		// correctly flag normal vehicles
		case a == r && a == false:
			tn++
		// incorrectly flag as normal vehicles
		case a != r && a == true:
			fn++
		// incorrectly flags as misbehaving vehicles
		case a != r && a == false:
			fp++
		}
	}
	res.Tp, res.Tn, res.Fn, res.Fp = tp, tn, fn, fp

	// advanced metrics
	recall := tp / (tp + fn)
	precision := tp / (tp + fp)
	f1ssimulator := 2 * recall * precision / (recall + precision)
	acc := (tp + tn) / (tp + tn + fp + fn)
	res.Recall, res.Precision, res.F1Score, res.Acc = recall, precision, f1ssimulator, acc

	return res
}

func Run(ctx context.Context) int {
	logutil.GetLogger(PackageName).Debugf("[Run] start statistics service")
	return 0
}

func Done() int {
	return 0
}
