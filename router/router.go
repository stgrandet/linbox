package router

import (
	gonsq "github.com/bitly/go-nsq"
	"github.com/msbranco/goconfig"
	logger "github.com/cihub/seelog"
)

const (
	// TODO
	// 暂时使用配置文件，之后改为动态配置
	configFileName string = "nsq.cfg"
	nsqlookupSection string = "nsqlookup"
)

var (
	nsqLookupAddrs []string
	consumerPool []*gonsq.Consumer
)

func StartRouter() {

}

func StopRouter() {

}

func initConsumer() {
	configs, err := goconfig.ReadConfigFile(configFileName)

	if err != nil {
		logger.Criticalf("Can not read nsqlookup config from %s. Error: %s", configFileName, err)
		panic(err)
	}

	options, err := configs.GetOptions(nsqlookupSection)

	if err != nil {
		logger.Criticalf("Can not find configs for nsqlookup in %s. Error: %s", configFileName, err)
		panic(err)
	}


}


