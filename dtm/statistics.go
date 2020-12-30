package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/statistics"
)

func (session *DTMLogicSession) genStatistics(ctx context.Context, epoch uint64) {
	logutil.LoggerList["dtm"].Debugf("[genStatistics] epoch %v", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["dtm"].Debugf("[genStatistics] context canceled")
		return
	default:
		// iterate through every experiment's data storage
		for expName := range *session.Config {
			head := (*session.TrustValueStorageHead)[expName]

			go func(expName string) {
				// get the head of the storage
				headBlock := head.GetHeadBlock()
				ep, _, bmap := headBlock.GetTrustValueList()
				if ep != epoch {
					logutil.LoggerList["dtm"].Debugf("[flagMisbehavingVehicles] epoch mismatch! ep %v, epoch %v", ep, epoch)
					return
				}

				// generate statistics
				bundle := statistics.GenStatisticsForEpoch(epoch, session.MisbehavingVehicleBitMap, bmap)
				headBlock.SetStatistics(bundle)

				// debug
				logutil.LoggerList["dtm"].Infof("[genStatistics] epoch %v: exp %v, tp %v, tn %v, fp %v, fn %v, recall %v, precision %v, f1 %v, acc %v",
					epoch, expName, bundle.TP, bundle.TN, bundle.FP, bundle.FN, bundle.Recall, bundle.Precision, bundle.F1ssimulator, bundle.ACC)

			}(expName)
		}
	}
}
