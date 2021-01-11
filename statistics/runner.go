package statistics

import (
	"context"
	"fmt"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"os"
	"reflect"
	"strings"
)

func (session *StatisticsSession) Init() {
	dir := fmt.Sprintf(session.Config.RootPath + session.Config.Dir)

	// create the dir to hold log files
	err := os.Mkdir(dir, 0777)
	if err != nil {
		logutil.LoggerList["statistics"].Fatalf("[Init] failed")
	}

	// create each log file
	for _, mtype := range session.Config.MetricsType {
		f, err := os.Create(dir + MetricsNameMapping[mtype])
		if err != nil {
			logutil.LoggerList["statistics"].Fatalf("[Init] failed, %v", err)
		}
		session.FileDescriptors[mtype] = f
	}
}

func (session *StatisticsSession) Done() {
	for _, f := range session.FileDescriptors {
		f.Close()
	}
}

func (session *StatisticsSession) logData(data *pb.StatisticsBundle) {

	for _, mType := range session.Config.MetricsType {
		mName := MetricsNameMapping[mType]
		f := session.FileDescriptors[mType]

		// iterate through the statistics bundle
		resArray := make([]interface{}, len(session.ExperimentMapping))
		epoch, bundle := data.Epoch, data.Bundle
		resArray[0] = epoch

		for _, exp := range bundle {
			index := session.ExperimentMapping[exp.Name]
			resArray[index] = reflect.ValueOf(exp).FieldByName(mName)
		}

		// write the data to the log file
		if _, err := f.WriteString(strings.Trim(fmt.Sprintf("%v", resArray), "[]") + "\n"); err != nil {
			logutil.LoggerList["statistics"].Fatalf("[logData] failed, %v", err)
		}

	}
}

func (session *StatisticsSession) Run(ctx context.Context) {
	logutil.LoggerList["statistics"].Debugf("[Run] start!")

	// init the experiment log files

	// after initialization is finished, waiting for the communication from the simulator
	for {
		select {
		case <-ctx.Done():
			session.Done()
			logutil.LoggerList["statistics"].Fatalf("[Run] context canceled")
		case v := <-session.ChanDTM:
			// using reflect to detect what is being passed to the dtm runner
			switch v.(type) {
			case pb.StatisticsBundle: // signal for epoch
				// unpack
				pack := v.(*pb.StatisticsBundle)

				session.processRawData(ctx, pack)
				session.writeToFile(ctx)
			}
		}
	}
}
