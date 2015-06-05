package connector

import (
	"errors"

	logger "github.com/cihub/seelog"

	. "linbox/messages"
	"nsq"
)

type MqService struct {
}

func (*MqService) InitService() {
	nsq.InitProducerPool()
}

func (*MqService) HandleReceivingMsg(msgType uint16, buf []byte) error {
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
		err = errors.New("Unknow msgType: " + msgType)
	}

	return err
}

func (*MqService) HandleSendingMsg(uint64, chan<- []byte, chan<- bool) {

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
