package logutil

import log "github.com/sirupsen/logrus"

var LoggerList = make(map[string]*log.Entry)
var PackageNameList = []string{
	"simulator",
	"main",
	"simmap",
	"vehicle",
	"rsu",
	"service",
	"statistics",
	"dtm",
}

func InitLogger() {
	for _, v := range PackageNameList {
		RegisterLogger(v)
	}
}

func RegisterLogger(prefix string) {
	log.SetLevel(log.InfoLevel)
	fields := log.Fields{
		"package": prefix,
	}
	LoggerList[prefix] = log.WithFields(fields)
}

func GetLogger(prefix string) *log.Entry {
	return LoggerList[prefix]
}
