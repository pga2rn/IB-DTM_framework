package logutil

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

var LoggerList = make(map[string]*log.Entry)
var ServiceList interface{}
var mu = sync.RWMutex{}

func SetLevel(level log.Level) {
	log.SetLevel(level)
}

func SetServiceList(serviceList interface{}) {
	ServiceList = serviceList
}

func RegisterLogger(prefix string) {
	mu.Lock()
	defer mu.Unlock()
	fields := log.Fields{
		"package": prefix,
	}
	LoggerList[prefix] = log.WithFields(fields)
}

func GetLogger(prefix string) *log.Entry {
	if _, ok := LoggerList[prefix]; !ok {
		RegisterLogger(prefix)
	}

	mu.RLock()
	defer mu.RUnlock()
	return LoggerList[prefix]
}
