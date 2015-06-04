package nsq

import (
	"sync"
	"math/rand"

	logger "github.com/cihub/seelog"
	gonsq "github.com/bitly/go-nsq"
	"github.com/msbranco/goconfig"
	"time"
	"errors"
)

const (
	// TODO
	// 目前使用配置文件进行配置，之后改为动态配置
	configFileName string = "nsq.cfg"
	nsqdConfigSection string = "nsqd"

	autoReconnectInterval time.Duration = time.Second * 30
)

var (
	once sync.Once
	rwlock sync.RWMutex

	pool map[string]*gonsq.Producer
	lostConns []string
)

func InitProducerPool() {
	once.Do(initPool)
}

func GetProducer() (*gonsq.Producer, error) {
	defer rwlock.RUnlock()

	rwlock.RLock()

	size := len(pool)
	if size == 0 {
		logger.Criticalf("No available connection for nsqds")
		return nil, errors.New("No find available connection for nsqds")
	}

	var producer *gonsq.Producer

	index := rand.Intn(size)
	for _, p := range pool {
		if index == 0 {
			producer = p
		}

		index--
	}

	return producer, nil
}

func MarkConnectFail(producer *gonsq.Producer) {
	defer rwlock.Unlock()

	rwlock.Lock()

	var addr string = ""

	for a, p := range pool {
		if producer == p {
			addr = a
		}
	}

	if addr == "" {
		logger.Errorf("MarkConnectioinFail can not find correct producer: %s", producer.String())
		return
	}

	logger.Errorf("nsqd %s lost connection.", addr)

	delete(pool, addr)
	lostConns = append(lostConns, addr)

}

func initPool() {
	configs, err := goconfig.ReadConfigFile(configFileName)

	if err != nil {
		logger.Criticalf("Can not read nsq configs from nsq.cfg. Error: %s", err)
		panic(err)
	}

	options, err := configs.GetOptions(nsqdConfigSection)

	if err != nil {
		logger.Criticalf("Can not read nsqd config in nsq.cfg. Error: $s", err)
		panic(err)
	}

	addrs := make([]string, 0, len(options))

	for _, option := range options {
		value, err := configs.GetString(nsqdConfigSection, option)

		if err != nil {
			logger.Errorf("Get error when reading section %s option %s in nsq.cfg. Error: %s", configFileName, option, err)
			continue
		}

		addrs = append(addrs, value)
	}

	if len(addrs) <= 0 {
		logger.Criticalf("Read 0 configs for nsqd address in nsq.cfg.")
		panic("Read 0 configs for nsqd address in nsq.cfg.")
	}

	pool = make(map[string]*gonsq.Producer)
	lostConns = make([]string, 0)

	for _, addr := range addrs {
		config := gonsq.NewConfig()
		producer, err := gonsq.NewProducer(addr, config)

		if err != nil {
			logger.Errorf("Can not create nsq producer for address: %s. Error: %s", addr, err)
			continue
		}

		err = producer.Ping()

		if err != nil {
			logger.Errorf("Can not connect to address %s. Error: %s", addr, err)
			lostConns = append(lostConns, addr)
		}

		pool[addr] = producer
	}

	go autoReconnect()
}

// 随机对连接失效的 nsqd 进行自动重连
// 每次只随机取出一个丢失的连接进行尝试
func autoReconnect() {
	for {
		time.Sleep(autoReconnectInterval)

		rwlock.Lock()

		count := len(lostConns)

		if count == 0 {
			rwlock.Unlock()
			continue
		}

		index := rand.Intn(count)
		addr := lostConns[index]

		logger.Info("Reconnecting nsqd %s", addr)

		producer, err := connectNsqd(addr)

		if err != nil {
			logger.Errorf("Reconnect nsqd %s error. Error: %s", addr, err)
			rwlock.Unlock()
			continue
		}

		lostConns = append(lostConns[:index], lostConns[(index+1):]...)
		pool[addr] = producer
		rwlock.Unlock()
	}
}

func connectNsqd(addr string) (*gonsq.Producer, error) {
	config := gonsq.NewConfig()
	producer, err := gonsq.NewProducer(addr, config)

	if err != nil {
		return nil, err
	}

	err = producer.Ping()

	if err != nil {
		return nil, err
	}

	return producer, nil
}