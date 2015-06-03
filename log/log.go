package log

/*
Use seelog as logger system.
Also Adapte seelog to glog interface, in order to use seelog in grpc package.
*/

import (
	"go/build"
	"sync"

	log "linbox/seelog"
)

const (
	DefSeelogCfgPath     string = "linbox/config"
	DefSeelogCfgFileName string = "seelog.xml"
)

var (
	logConfigLocker sync.Mutex = sync.Mutex{}
)

func InitDefaultLogger() {
	sync.Once.Do(readDefaultLogConfig)
}

func InitLoggerByConfig(fileName string) {
	defer func() {
		logConfigLocker.Unlock()
	}()

	logConfigLocker.Lock()

	logger, err := log.LoggerFromConfigAsFile(fileName)
	if err != nil {
		log.Errorf("Can not create log from config file. FileName: %s", fileName)
		return
	}

	log.ReplaceLogger(logger)
}

func readDefaultLogConfig() {
	cfgPackage, err := build.Default.Import(DefSeelogCfgPath, "", build.FindOnly)
	if err != nil {
		log.Errorf("Can not find default log config package for zkkeeper. Package path: %s", DefSeelogCfgPath)
		return
	}

	cfgName := cfgPackage.Dir + "/" + DefSeelogCfgFileName

	logger, err := log.LoggerFromConfigAsFile(cfgName)
	if err != nil {
		log.Errorf("Can not create logger from default log config for zkkeeper. Log file full path: %s", cfgName)
		return
	}

	log.ReplaceLogger(logger)
}