package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/statistics"
	"sync"
)

func (session *DTMLogicSession) genStatistics(ctx context.Context, epoch uint32) {
	logutil.LoggerList["dtm"].Debugf("[genStatistics] epoch %v", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["dtm"].Debugf("[genStatistics] context canceled")
		return
	default:
		wg := sync.WaitGroup{}

		// iterate through every experiment's data storage
		for expName, exp := range *session.ExpConfig {
			head := (*session.TrustValueStorageHead)[expName]

			wg.Add(1)
			go func(expName string, exp *config.ExperimentConfig) {
				// get the head of the storage
				headBlock := head.GetHeadBlock()
				ep, _, bmap := headBlock.GetTrustValueList()
				if ep != epoch {
					logutil.LoggerList["dtm"].Debugf("[flagMisbehavingVehicles] epoch mismatch! ep %v, epoch %v", ep, epoch)
					return
				}

				// generate statistics
				pack := statistics.GenStatisticsForEpoch(epoch, session.MisbehavingVehicleBitMap, bmap)
				pack.Epoch, pack.Name, pack.Type = epoch, expName, exp.Type
				headBlock.SetStatistics(pack)

				wg.Done()
				// debug
				logutil.LoggerList["dtm"].Infof("epoch %v, exp %v, tp %v, tn %v, fp %v, fn %v, recall %v, precision %v, f1 %v, acc %v",
					epoch, expName, pack.Tp, pack.Tn, pack.Fp, pack.Fn, pack.Recall, pack.Precision, pack.F1Score, pack.Acc)
			}(expName, exp)
		}

		// wait for all jobs to be done
		wg.Wait()
	}
}
