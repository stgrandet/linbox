package service

type MqService struct {

}

func (*MqService) RegisterTopic() {

}

func (*MqService) SendToTopic() {

}
func (*MqService) RegisterToChannel() {

}
func (*MqService)HandleReceivingMsg(msgType uint16, buf []byte) {

}
func (*MqService) HandleSendingMsg(uint64, chan<- []byte, chan<- bool) {

}

