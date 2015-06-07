package connector

import (
	logger "github.com/cihub/seelog"

	"linbox/connector/nsq"
	. "linbox/messages"
)

type MqService struct {
}

func (*MqService) InitService() {
	nsq.InitProducerPool()
}

func (s *MqService) HandleReceivingMsg(msgType uint16, msg []byte) {
	switch msgType {
	case SYNC_UNREAD_REQUEST:
		s.handleSyncUnreadRequest(msg)
	case READ_ACK_REQUEST:
		s.handleReadAckRequest(msg)
	case PULL_OLD_MSG_REQUEST:
		s.handlePullOldMsgRequest(msg)
	case SEND_MSG_REQUEST:
		s.handleSendMsgRequest(msg)
	default:
		logger.Errorf("Unknown msgType: %d", msgType)
	}
}

func (*MqService) HandleSendingMsg(userId uint64, listenChannel chan<- []byte, quit <-chan bool) {

}

func (*MqService) handleSyncUnreadRequest(msg []byte) {
	err := publishMsg(TOPIC_SYNC_UNREAD_REQUEST, msg)

	// TODO
	// 增加错误处理
	if err != nil {

	}
}

func (*MqService) handleReadAckRequest(msg []byte) {
	err := publishMsg(TOPIC_READ_ACK_REQUEST, msg)

	// TODO
	// 增加错误处理
	if err != nil {

	}
}

func (*MqService) handlePullOldMsgRequest(msg []byte) {
	err := publishMsg(TOPIC_PULL_OLD_MSG_REQUEST, msg)

	// TODO
	// 增加错误处理
	if err != nil {

	}
}

func (*MqService) handleSendMsgRequest(msg []byte) {
	err := publishMsg(TOPIC_SEND_MSG_REQUEST, msg)

	// TODO
	// 增加错误处理
	if err != nil {

	}
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
