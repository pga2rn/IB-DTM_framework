package statistics

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

type Statistics struct {
}

func Run(ctx context.Context) int {
	logutil.LoggerList["statistics"].Debugf("[Run] start statistics service")
	return 0
}

func Done() int {
	return 0
}
