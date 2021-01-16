package logutil

import log "github.com/sirupsen/logrus"

var LoggerList = make(map[string]*log.Entry)
var LogLevel = log.DebugLevel
var PackageNameList = []string{
	"simulator",
	"main",
	"simmap",
	"vehicle",
	"rsu",
	"service",
	"statistics",
	"dtm",
	"ib-dtm",
	"rpc",
}

func SetLevel(level log.Level) {
	LogLevel = level
}

func InitLogger() {
	log.SetLevel(LogLevel)
	for _, v := range PackageNameList {
		RegisterLogger(v)
	}
}

func RegisterLogger(prefix string) {
	fields := log.Fields{
		"package": prefix,
	}
	LoggerList[prefix] = log.WithFields(fields)
}

func GetLogger(prefix string) *log.Entry {
	return LoggerList[prefix]
}
