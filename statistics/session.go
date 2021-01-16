package statistics

import (
	"context"
	"fmt"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"os"
	"reflect"
	"strings"
)

var PackageName = "statistics"

type StatisticsSession struct {
	Config *config.StatisticsConfig

	// mapping experiment with index
	ExperimentMapping map[string]int
	FileDescriptors   map[pb.MetricsType]*os.File

	// map[metric_type][experimentIndex]results
	Epoch         uint32
	MetricsBundle map[pb.MetricsType]*[]float32

	ChanDTM chan interface{}
}

var MetricsNameMapping = map[pb.MetricsType]string{
	pb.MetricsType_TruePositive:  "Tp",
	pb.MetricsType_TrueNegative:  "Tn",
	pb.MetricsType_FalsePositive: "Fp",
	pb.MetricsType_FalseNegative: "Fn",
	pb.MetricsType_Recall:        "Recall",
	pb.MetricsType_Precision:     "Precision",
	pb.MetricsType_F1Score:       "F1Score",
	pb.MetricsType_Accuracy:      "Acc",
}

var MetricsTypeMapping = map[string]pb.MetricsType{
	"Tp":        pb.MetricsType_TruePositive,
	"Tn":        pb.MetricsType_TrueNegative,
	"Fp":        pb.MetricsType_FalsePositive,
	"Fn":        pb.MetricsType_FalseNegative,
	"Recall":    pb.MetricsType_Recall,
	"Precision": pb.MetricsType_Precision,
	"F1Score":   pb.MetricsType_F1Score,
	"Acc":       pb.MetricsType_Accuracy,
}

func PrepareStatisticsSession(cfg *config.StatisticsConfig, expList *map[string]*config.ExperimentConfig) *StatisticsSession {
	session := &StatisticsSession{
		Config:            cfg,
		ExperimentMapping: make(map[string]int),
	}

	// register experiment
	count := 0
	for expName := range *expList {
		session.ExperimentMapping[expName] = count
		count++
	}

	return session
}

func (session *StatisticsSession) processRawData(ctx context.Context, pack *pb.StatisticsBundle) {
	session.MetricsBundle = make(map[pb.MetricsType]*[]float32)
	for metricsType := range MetricsNameMapping {
		tmp := make([]float32, len(session.ExperimentMapping))
		session.MetricsBundle[metricsType] = &tmp
	}

	if session.Epoch < pack.Epoch {
		session.Epoch = pack.Epoch
	} else {
		logutil.GetLogger(PackageName).Debugf("[processRawdata] received older data")
		return
	}

	for _, dataBundle := range (*pack).Bundle {
		expIndex := session.ExperimentMapping[dataBundle.Name]

		val := reflect.ValueOf(dataBundle).Elem()
		for i := 0; i < val.NumField(); i++ {
			mType, ok := MetricsTypeMapping[val.Type().Field(i).Name]
			if !ok { // the data field is not metric
				continue
			}

			// extract the value from the bundle
			(*session.MetricsBundle[mType])[expIndex] = val.Field(i).Interface().(float32)
		}
	}
}

func (session *StatisticsSession) writeToFile(ctx context.Context) {
	epoch := session.Epoch
	for metricType, metricArray := range session.MetricsBundle {
		f := session.FileDescriptors[metricType]
		result := fmt.Sprintf("epoch %v:", epoch) + strings.Trim(fmt.Sprintf("%v", metricArray), "[]")

		_, err := f.WriteString(result)
		if err != nil {
			logutil.GetLogger(PackageName).Debugf("[writeToFile] %v", err)
		}
	}
}
