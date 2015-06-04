package service

type ConnectorService interface {
	registerTopic()
	sendToTopic()
	registerToChannel()
	handleReceivingMsg()
	handleSendingMsg()
}



