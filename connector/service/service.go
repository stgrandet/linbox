package service

type ConnectorService interface {
	RegisterTopic()
	SendToTopic()
	RegisterToChannel()
	HandleReceivingMsg( uint16,  []byte)
	HandleSendingMsg(uint64, chan<- []byte, chan<- bool)
}



