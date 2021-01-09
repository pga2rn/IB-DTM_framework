package config

import (
	"fmt"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"time"
)

type StatisticsConfig struct {
	RootPath    string
	Dir         string
	MetricsType []pb.MetricsType
}

func GenStatisticsConfig() *StatisticsConfig {
	return &StatisticsConfig{
		RootPath: "D:/EXP/",
		Dir:      fmt.Sprintf("%v", time.Now().Unix()),
		MetricsType: []pb.MetricsType{
			pb.MetricsType_TruePositive,
			pb.MetricsType_TrueNegative,
			pb.MetricsType_FalsePositive,
			pb.MetricsType_FalseNegative,
			pb.MetricsType_Recall,
			pb.MetricsType_Precision,
			pb.MetricsType_F1Score,
			pb.MetricsType_Accuracy,
		},
	}
}
