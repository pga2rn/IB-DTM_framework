package statistics

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

type Statistics struct {
}

func Run(ctx context.Context) {
	logutil.LoggerList["statistics"].Debugf("[Run] start statistics service")
}
