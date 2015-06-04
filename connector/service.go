package connector

type ConnectorService interface {
	InitService()
	HandleReceivingMsg( uint16,  []byte)
	HandleSendingMsg(uint64, chan<- []byte, chan<- bool)
}



