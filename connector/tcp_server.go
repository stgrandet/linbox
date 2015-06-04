package connector

import (
	"io"
	"net"
	"time"
	"linbox/connector/service"

	"encoding/binary"
	logger "linbox/seelog"
)

const (
	readTimeoutInHour int   = 4
	writeTimeoutInSecond int = 10
	bufferSize        int64 = 1024 * 1024
	sendingChannelSize int = 100
)

var (
	  connService service.ConnectorService
)

func StartTcpServer(host, port string, service *service.ConnectorService) {

	if service == nil {
		connService = &service.MqService{}
	} else {
		connService = service
	}

	addr, err := net.ResolveTCPAddr("tcp", ":9000")
	if err != nil {
		logger.Errorf("Can not resolve tcp address for server. host: %s. port: %s. address string: ")
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Criticalf("Can not start tcp server on address: %s:%s. Error: %s", addr.IP, addr.Port, err)
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logger.Errorf("Create tcp connection error. Err: %s", err)
			continue
		}

		handleConnection(conn)
	}
}

func StopTcpServer() {

}

func handleConnection(conn *net.TCPConn) {
	conn.SetKeepAlive(true)

	userid, err := authenticate(conn)
	if err != nil {
		logger.Errorf("Authenticate Information Fail")
		conn.Close()
		return
	}

	go handleReceivingMsg(conn)
	go handleSendingMsg(conn, userid)
}

func authenticate(conn *net.TCPConn) (uint64, error) {
	return 1000, nil
}

func handleReceivingMsg(conn *net.TCPConn) {
	defer func() {
		conn.Close()
	}()

	for {
		now := time.Now()
		timeout := now.Add(time.Hour * readTimeoutInHour)
		conn.SetReadDeadline(timeout)

		msgTypeByte := make([]byte, 1)
		_, err := io.ReadFull(conn, msgTypeByte)
		if err != nil {
			logger.Errorf("Connection Read Msg Type Error. Error: %s.", err)
			break
		}

		msgType := int(binary.BigEndian.Uint16(msgTypeByte))

		msgLenByte := make([]byte, 4)
		_, err = io.ReadFull(conn, msgLenByte)
		if err != nil {
			logger.Errorf("Connection Read Msg Length Error. Error: %s.", err)
			break
		}

		msgLen := int(binary.BigEndian.Uint32(msgLenByte))

		buf := make([]byte, int64(msgLen))
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			logger.Errorf("Connection Read Msg Content Error. Error: %s", err)
			break
		}

		connService.HandleReceivingMsg(msgType, msgLen)
	}
}

func handleSendingMsg(conn *net.TCPConn, userid uint64) {
	defer conn.Close()

	channel := make(chan []byte, sendingChannelSize)
	quit := make(chan bool)

	defer func(){
		quit<-true
	}()

	connService.HandleSendingMsg(userid, channel, quit)

	for {
		message, more := <-channel

		if !more {
			logger.Errorf("The sending channel is closed by connector service. Close connection")
			break
		}

		now := time.Now()
		timeout := now.Add(time.Second * writeTimeoutInSecond)
		conn.SetWriteDeadline(timeout)

		_, err := conn.Write(message)

		if err != nil {
			logger.Errorf("Sending messages for user %d Error. Close the connection. Error: %s", userid, err)
			break
		}
	}
}

