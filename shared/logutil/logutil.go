package logutil

import log "github.com/sirupsen/logrus"

var LoggerList = make(map[string]*log.Entry)
var PackageNameList = []string{
	"core",
	"main",
	"simmap",
	"vehicle",
	"rsu",
}

func InitLogger(){
	for _, v := range PackageNameList{
		RegisterLogger(v)
	}
}

func RegisterLogger(prefix string) {
	log.SetLevel(log.DebugLevel)
	LoggerList[prefix] = log.WithField("prefix", prefix)
}

func GetLogger(prefix string) *log.Entry {
	return LoggerList[prefix]
}
