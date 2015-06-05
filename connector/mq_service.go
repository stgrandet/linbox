package connector

import (
	"errors"

	logger "github.com/cihub/seelog"

	. "linbox/messages"
	"linbox/connector/nsq"
)

type MqService struct {
}

func (*MqService) InitService() {
	nsq.InitProducerPool()
}

// TODO
// 这里需要添加错误处理
// 如果信息无法正确下发到 MQ，则直接向客户端发送错误信息，以使客户端能够显示
func (*MqService) HandleReceivingMsg(msgType uint16, buf []byte) {
	var err error = nil

	switch msgType {
	case SYNC_UNREAD_REQUEST:
		err = publishMsg(TOPIC_SYNC_UNREAD_REQUEST, buf)
	case READ_ACK_REQUEST:
		err = publishMsg(TOPIC_READ_ACK_REQUEST, buf)
	case PULL_OLD_MSG_REQUEST:
		err = publishMsg(TOPIC_PULL_OLD_MSG_REQUEST, buf)
	case SEND_MSG_REQUEST:
		err = publishMsg(TOPIC_SYNC_UNREAD_REQUEST, buf)
	default:
		logger.Errorf("Unknown msgType: %d", msgType)
		err = errors.New("Unknow msgType ")
	}

	if err != nil {

	}
}

func (*MqService) HandleSendingMsg(userId uint64, listenChannel chan<- []byte, quit <-chan bool) {

}

func publishMsg(topic string, msg []byte) error {
	producer, err := nsq.GetProducer()

	if err != nil {
		logger.Criticalf("Can not get available producer from nsqd pool. Error: %s", err)
		return err
	}

	err = producer.Publish(topic, msg)

	if err != nil {
		logger.Criticalf("Publish topic %s error. Producer: %s. Error: %s", topic, producer.String(), err)
		return err
	}

	return nil
}
