package statistics

import (
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"os"
)

type StatisticsSession struct {
	Config *config.StatisticsConfig

	// mapping experiment with index
	ExperimentMapping map[string]int
	FileDescriptors   map[pb.MetricsType]*os.File

	ChanDTM chan interface{}
}

var MetricsName = map[pb.MetricsType]string{
	pb.MetricsType_TruePositive:  "Tp",
	pb.MetricsType_TrueNegative:  "Tn",
	pb.MetricsType_FalsePositive: "Fp",
	pb.MetricsType_FalseNegative: "Fn",
	pb.MetricsType_Recall:        "Recall",
	pb.MetricsType_Precision:     "Precision",
	pb.MetricsType_F1Score:       "F1Score",
	pb.MetricsType_Accuracy:      "Acc",
}
