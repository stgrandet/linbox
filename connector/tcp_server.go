package connector

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"encoding/json"
	logger "github.com/cihub/seelog"
	. "linbox/messages"
)

const (
	readTimeoutInHour           int   = 4
	writeTimeoutInSecond        int   = 10
	bufferSize                  int64 = 1024 * 1024
	sendingChannelSize          int   = 100
	authenticateTimeoutInSecond int   = 60
)

var (
	connService ConnectorService
)

func StartTcpServer(host, port string, service ConnectorService) {

	if service == nil {
		connService = &MqService{}
	} else {
		connService = service
	}
	connService.InitService()

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

	userid, err := handleAuth(conn)
	if err != nil {
		logger.Errorf("Authenticate Information Fail. Userid: %d. Error: %s", userid, err)
		conn.Close()
		return
	}

	go handleReceivingMsg(conn)
	go handleSendingMsg(conn, userid)
}

func handleAuth(conn *net.TCPConn) (userid uint64, err error) {
	userid, err = authenticate(conn)
	if err != nil {
		logger.Errorf("Authenticate Information Fail")
		sendFailAuthResponse(conn, userid, ERROR_AUTH_FAIL, err.Error())
		return
	}

	err = sendSuccessAuthResponse(conn, userid)

	return
}

func authenticate(conn *net.TCPConn) (userid uint64, err error) {
	buf := make([]byte, 1000)

	now := time.Now()
	timeout := now.Add(time.Second * time.Duration(authenticateTimeoutInSecond))

	conn.SetReadDeadline(timeout)

	length, err := conn.Read(buf)
	if err != nil {
		logger.Errorf("Read authenticate Error: %s", err)
		return 0, err
	}

	buf = buf[0:length]

	conn.SetReadDeadline(time.Time{})

	authRequest := &AuthRequest{}
	err = json.Unmarshal(buf, authRequest)
	if err != nil {
		logger.Errorf("Convert auth info from json error. Json: %s. Error: %s.", buf, err)
		return 0, err
	}

	return authRequest.UserId, nil
}

func sendSuccessAuthResponse(conn *net.TCPConn, userId uint64) error {
	message, err := BuildAuthResponseBuf(userId, SUCCESS_RESPONSE, "")
	if err != nil {
		logger.Errorf("Build AuthResponseBuf error. UserId: %d. Error: err", userId, err)
		return err
	}

	return sendAuthResponse(conn, message)
}

func sendFailAuthResponse(conn *net.TCPConn, userId uint64, errCode uint32, errMsg string) {
	message, err := BuildAuthResponseBuf(userId, errCode, errMsg)
	if err != nil {
		logger.Errorf("Build AuthResponseBuf error. UserId: %d. Error: err", userId, err)
		return err
	}

	return sendAuthResponse(conn, message)
}

func sendAuthResponse(conn *net.TCPConn, msg []byte) error {
	now := time.Now()
	timeout := now.Add(time.Second * time.Duration(60))

	conn.SetWriteDeadline(timeout)

	_, err := conn.Write(msg)
	if err != nil {
		logger.Errorf("Send AuthResponse Fail. AuthResponseBuf: %+v. Error: %s", msg, err)
		return err
	}

	return nil
}

func handleReceivingMsg(conn *net.TCPConn) {
	defer func() {
		conn.Close()
	}()

	for {
		now := time.Now()
		timeout := now.Add(time.Hour * time.Duration(readTimeoutInHour))
		conn.SetReadDeadline(timeout)

		msgTypeByte := make([]byte, 1)
		_, err := io.ReadFull(conn, msgTypeByte)
		if err != nil {
			logger.Errorf("Connection Read Msg Type Error. Error: %s.", err)
			break
		}

		msgTypeByte = append(make([]byte, 1), msgTypeByte[0])
		msgType := uint16(binary.BigEndian.Uint16(msgTypeByte))

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

		connService.HandleReceivingMsg(msgType, buf)

	}
}

func handleSendingMsg(conn *net.TCPConn, userid uint64) {
	defer conn.Close()

	channel := make(chan []byte, sendingChannelSize)
	quit := make(chan bool, 1)

	defer func() {
		quit <- true
		close(quit)
	}()

	connService.HandleSendingMsg(userid, channel, quit)

	for {
		message, more := <-channel

		if !more {
			logger.Errorf("The sending channel is closed by connector service. Close connection")
			break
		}

		now := time.Now()
		timeout := now.Add(time.Second * time.Duration(writeTimeoutInSecond))
		conn.SetWriteDeadline(timeout)

		_, err := conn.Write(message)

		if err != nil {
			logger.Errorf("Sending messages for user %d Error. Close the connection. Error: %s", userid, err)
			break
		}
	}
}
