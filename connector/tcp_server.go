package connector

import (
	"net"
	"strings"
	"time"
	"io"

	logger "linbox/seelog"
	"encoding/binary"
)

const (
	readTimeoutInHour int = 4
	bufferSize int64 = 1024 * 1024
)

func StartTcpServer(host, port string) {
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

		conn.SetKeepAlive(true)

		go handleReceivingMsg(conn)
		go handleSendingMsg(conn)
	}
}

func StopTcpServer() {

}

func handleReceivingMsg(conn net.Conn) {
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

		// TODO
		// Add all possible type here
		switch msgType {
			case SYNC_UNREAD_REQUEST:

			default:

		}

	}

	conn.Close()
}

func handleSendingMsg(conn net.Conn) {
	for {
		break
	}

	conn.Close()
}
