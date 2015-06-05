package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	logger "github.com/cihub/seelog"
	"linbox/connector"
	. "linbox/messages"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


var (
	inBox       map[uint64]map[string]UnreadMsg // 未读信息收件箱
	messageBox  map[uint64][][]byte             // 发送信息箱
	msgIdRecord map[string]uint64

	inBoxRwMutex      sync.RWMutex // 发件箱读写锁
	messageBoxRWMutex sync.RWMutex // 消息读写锁

)


func init() {
	inBox = make(map[uint64]map[string]UnreadMsg)
	messageBox = make(map[uint64][][]byte)
	msgIdRecord = make(map[string]uint64)
}

type testConnService struct {
	msgChan chan []byte
}



func (s *testConnService) InitService() {
	s.msgChan = make(chan []byte, 100)
}

func (s *testConnService) HandleReceivingMsg(msgType uint16, message []byte) {

	switch msgType {
	case SYNC_UNREAD_REQUEST:
		logger.Infof("Received SYNC_UNREAD_REQUEST message %s", message)

		syncUnreadRequest := &SyncUnreadRequest{}
		err := json.Unmarshal(message, syncUnreadRequest)
		if err != nil {
			logger.Errorf("Can not parse SyncUnreadRequest from json. Error: %s", err)
			return
		}

		rid := syncUnreadRequest.Rid
		userId := syncUnreadRequest.UserId
		sessionKey := syncUnreadRequest.SessionKey

		if sessionKey != "" {
			defer inBoxRwMutex.RUnlock()
			inBoxRwMutex.RLock()

			allMessages, ok := inBox[userId]
			if !ok || len(allMessages) < 0 {
				s.sendEmptySyncUnreadResponse(rid, userId)
				return
			}

			unReadMsg, ok := allMessages[sessionKey]
			if !ok {
				s.sendEmptySyncUnreadResponse(rid, userId)
				return
			}

			s.sendSyncUnreadResponse(rid, userId, unReadMsg)
			return
		} else {

			defer inBoxRwMutex.RUnlock()
			inBoxRwMutex.RLock()

			// 拉取所有的未读信息
			allMessages, ok := inBox[userId]
			if !ok || len(allMessages) < 0 {
				s.sendEmptySyncUnreadResponse(rid, userId)
				return
			} else {
				s.sendSyncUnreadResponse(rid, userId, allMessages[sessionKey])
				return
			}
		}
	case READ_ACK_REQUEST:
		logger.Infof("Received READ_ACK_REQUEST message %s", message)

		readAckRequest := &ReadAckRequest{}
		err := json.Unmarshal(message, readAckRequest)
		if err != nil {
			logger.Errorf("Can not parse ReadAckRequest from json. Error: %s", err)
			return
		}

		rid := readAckRequest.Rid
		userId := readAckRequest.UserId
		sessionKey := readAckRequest.SessionKey
		msgId := readAckRequest.MsgId

		defer inBoxRwMutex.Unlock()
		inBoxRwMutex.Lock()

		allMessages, ok := inBox[userId]
		if !ok || len(allMessages) < 0 {
			s.sendErrorReadAckResponse(rid, userId, sessionKey, ERROR_EMPTY_INBOX, "No messages found for user")
			return
		}

		unReadMsg, ok := allMessages[sessionKey]
		if !ok {
			s.sendErrorReadAckResponse(rid, userId, sessionKey, ERROR_EMPTY_INBOX, "No messages found for session key")
			return
		}

		if unReadMsg.MsgId <= msgId {
			delete(allMessages, sessionKey)

			if len(allMessages) > 0 {
				inBox[userId] = allMessages
			} else {
				delete(inBox, userId)
			}

		} else {
			unReadMsg.Count = unReadMsg.MsgId - msgId
			allMessages[sessionKey] = unReadMsg
		}

		s.sendReadAckResponse(rid, userId, sessionKey)

	case PULL_OLD_MSG_REQUEST:
		logger.Infof("Received PULL_OLD_MSG_REQUEST message %s", message)

		pullOldRequest := &PullOldMsgRequest{}
		err := json.Unmarshal(message, pullOldRequest)
		if err != nil {
			logger.Errorf("Can not parse PullOldMsgRequest from json. Error: %s", err)
			return
		}

		rid := pullOldRequest.Rid
		userId := pullOldRequest.UserId
		remoteId := pullOldRequest.RemoteId
		maxMsgId := pullOldRequest.MaxMsgId
		limit := pullOldRequest.Limit

		for i := uint64(0); i < limit; i++ {
			msg := Message{}

			if i%2 != 0 {
				msg.From = userId
				msg.To = remoteId
			} else {
				msg.From = remoteId
				msg.To = userId
			}

			text := strconv.FormatUint(i, 10)

			msg.MsgId = maxMsgId - (limit - i)
			msg.Body = Content{
				Minetype: "plain/text",
				Text:     text,
				MediaId:  0,
			}
			msg.SentTime = uint64(time.Now().Nanosecond())

			s.sendPullOldMsgResponse(rid, userId, msg)
		}

	case SEND_MSG_REQUEST:
		logger.Infof("Received SEND_MSG_REQUEST message %s", message)

		sendMsgRequest := &SendMsgRequest{}
		err := json.Unmarshal(message, sendMsgRequest)
		if err != nil {
			logger.Errorf("Can not parse SendMsgRequest from json. Error: %s", err)
			return
		}

		from := sendMsgRequest.From
		to := sendMsgRequest.To
		msg := sendMsgRequest.Msg

		sessionKey := createSessionKey(from, to)

		var shouldPushNew = false

		defer inBoxRwMutex.Unlock()
		inBoxRwMutex.Lock()

		var msgId uint64

		if currentMsgID, ok := msgIdRecord[sessionKey]; ok {
			msgId = currentMsgID + 1
		} else {
			msgId = 1
		}

		msgIdRecord[sessionKey] = msgId

		allMessages, ok := inBox[to]
		if !ok {
			allMessages = make(map[string]UnreadMsg)
		}

		unreadMsg, ok := allMessages[sessionKey]
		if !ok {
			unreadMsg = UnreadMsg{
				SessionKey: sessionKey,
				MsgId:      0,
				Count:      1,
				Msg:        msg,
			}
		}

		if unreadMsg.MsgId < msgId {
			unreadMsg.Count += msgId - unreadMsg.MsgId
			unreadMsg.MsgId = msgId
			shouldPushNew = true
		}

		allMessages[sessionKey] = unreadMsg
		inBox[to] = allMessages

		if shouldPushNew {
			pushNew(to)
		}

	default:
		logger.Infof("Received other type message %s", message)
	}
}

func (s *testConnService) HandleSendingMsg(userid uint64, listenChannel chan<- []byte, quit <-chan bool) {

	go s.listenMessageBox(userid)

	go func(){
		defer close(listenChannel)

		for {
			select {
			case message, more := <-s.msgChan :
				if !more {
					break
				}

				listenChannel <- message

				logger.Infof("Send message: %s", message)
			case <-quit:
				logger.Info("Read quit info. Close connection")
				break
			}
		}
	}()
}

func (s *testConnService) listenMessageBox(userid uint64) {
	for {
		time.Sleep(time.Second * time.Duration(5))

		messageBoxRWMutex.Lock()

		if messages, ok := messageBox[userid]; ok {
			for _, message := range messages {
				logger.Infof("Before Put message from message box to channel. Message: %s", message)
				s.msgChan <- message
				logger.Infof("After Put message from message box to channel. Message: %s", message)
			}

			delete(messageBox, userid)
		}

		messageBoxRWMutex.Unlock()
	}
}

func getRemoteId(sessionKey string, userId uint64) (remoteId uint64, err error) {
	ids := strings.Split(sessionKey, "_")

	if len(ids) != 2 {
		return 0, errors.New("Session " + sessionKey + " Format is not correct")
	}

	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.Errorf("Get wrong id string %s from session key %s.", idStr, sessionKey)
			continue
		}

		if id != userId {
			return id, nil
		}
	}

	return 0, errors.New("Can not find remote id from session key " + sessionKey)
}

func (s *testConnService) sendEmptySyncUnreadResponse(rid, userId uint64) {
	response := SyncUnreadResponse{
		Rid:       rid,
		UserId:    userId,
		ErrorCode: SUCCESS_RESPONSE,
	}

	buf, err := json.Marshal(response)

	if err != nil {
		logger.Errorf("Can not convert SyncUnreadResponse %v to json", response)
		return
	}

	s.msgChan <- buf
}

func (s *testConnService) sendSyncUnreadResponse(rid, userId uint64, msgs ...UnreadMsg) {
	response := SyncUnreadResponse{
		Rid:       rid,
		UserId:    userId,
		Unread:    make([]UnreadMsg, 0),
		ErrorCode: SUCCESS_RESPONSE,
	}

	for _, msg := range msgs {
		response.Unread = append(response.Unread, msg)
	}

	buf, err := json.Marshal(response)

	if err != nil {
		logger.Errorf("Can not convert SyncUnreadResponse %v to json", response)
		return
	}

	s.msgChan <- buf
}

func (s *testConnService) sendErrorSyncUnreadResponse(rid, userId uint64, errCode uint32, errString string) {
	response := SyncUnreadResponse{
		Rid:       rid,
		UserId:    userId,
		ErrorCode: errCode,
		ErrorMsg:  errString,
	}

	buf, err := json.Marshal(response)

	if err != nil {
		logger.Errorf("Can not convert SyncUnreadResponse %v to json", response)
		return
	}

	s.msgChan <- buf
}

func (s *testConnService) sendErrorReadAckResponse(rid, userId uint64, sessionKey string, errCode uint32, errString string) {
	response := ReadAckResponse{
		Rid:        rid,
		UserId:     userId,
		SessionKey: sessionKey,
		ErrorCode:  errCode,
		ErrorMsg:   errString,
	}

	buf, err := json.Marshal(response)

	if err != nil {
		logger.Errorf("Can not convert ReadAckResponse %v to json", response)
		return
	}

	s.msgChan <- buf
}

func (s *testConnService) sendReadAckResponse(rid, userId uint64, sessionKey string) {
	response := ReadAckResponse{
		Rid:        rid,
		UserId:     userId,
		SessionKey: sessionKey,
		ErrorCode:  SUCCESS_RESPONSE,
	}

	buf, err := json.Marshal(response)

	if err != nil {
		logger.Errorf("Can not convert ReadAckResponse %v to json", response)
		return
	}

	s.msgChan <- buf
}

func createSessionKey(id1, id2 uint64) string {
	if id1 < id2 {
		return strconv.FormatUint(id1, 10) + strconv.FormatUint(id2, 10)
	} else {
		return strconv.FormatUint(id2, 10) + strconv.FormatUint(id1, 10)
	}
}

func (s *testConnService) sendPullOldMsgResponse(rid, userId uint64, msg Message) {
	response := PullOldMsgResponse{
		Rid:    rid,
		UserId: userId,
		Msg:    msg,
	}

	buf, err := json.Marshal(response)

	if err != nil {
		logger.Errorf("Can not convert PullOldMsgResponse %v to json", response)
		return
	}

	s.msgChan <- buf
}

func pushNew(remoteId uint64) {
	message := NewMessage{
		UserId: remoteId,
	}

	buf, err := json.Marshal(message)

	if err != nil {
		logger.Errorf("Can not convert NewMessage %v to json", message)
		return
	}

	messageBoxRWMutex.Lock()

	msgList, ok := messageBox[remoteId]
	if !ok {
		msgList = make([][]byte, 0)
	}

	msgList = append(msgList, buf)

	messageBox[remoteId] = msgList

	messageBoxRWMutex.Unlock()

}

func main() {
	param := os.Args[1]

	if param == "server" {
		startServer()
	} else if param == "client" {
		idString := os.Args[2]

		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			logger.Errorf("argument for userid error. userid: %s", idString)
			return
		}
		startClient(id)
	} else {
		logger.Error(`Use command: "test server" or "test client"`)
		return
	}
}

func startServer() {
	service := &testConnService{}
	service.InitService()

	connector.StartTcpServer("127.0.0.1", "9000", service)
}

func startClient(userId uint64) {
	addr, err := net.ResolveTCPAddr("tcp", ":9000")
	if err != nil {
		logger.Error("Tcp address can not be resolved")
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logger.Error("Can not link to server")
		return
	}
	defer conn.Close()

	auth := AuthRequest{
		UserId: userId,
	}

	// 建立连接后，需要首先发送包含 userid 的验证信息
	buf, err := json.Marshal(auth)
	if err != nil {
		logger.Error("AuthRequest to json error")
		return
	}

	conn.Write(buf)

	//	conn.SetKeepAlive(true)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			buf = make([]byte, 200)

			length, err := reader.Read(buf)
			if err != nil {
				logger.Infof("Read Console error: %s\n", err)
				break
			}

			buf = buf[0:length]

			block := strings.Split(string(buf), " ")
			command := strings.TrimSpace(block[0])

			switch command {
			case "SYNC_UNREAD_REQUEST":
				syncUnreadRequest := SyncUnreadRequest{
					Rid:    uint64(time.Now().Nanosecond()),
					UserId: userId,
				}

				buf, err := json.Marshal(syncUnreadRequest)
				if err != nil {
					logger.Errorf("SYNC_UNREAD_REQUEST parse json error")
					continue
				}

				length := len(buf)

				typeBuf := make([]byte, 2)
				lengthBuf := make([]byte, 4)

				binary.BigEndian.PutUint32(lengthBuf, uint32(length))
				binary.BigEndian.PutUint16(typeBuf, SYNC_UNREAD_REQUEST)

				buf = append(typeBuf[1:], append(lengthBuf, buf[0:]...)[0:]...)

				conn.Write(buf)

				logger.Infof("Send SYNC_UNREAD_REQUEST Message:\n%s", buf)

			case "READ_ACK_REQUEST":
				sessionKey := block[1]
				msgID, _ := strconv.ParseUint(block[2], 10, 64)

				readAckRequest := ReadAckRequest{
					Rid:        uint64(time.Now().Nanosecond()),
					UserId:     userId,
					SessionKey: sessionKey,
					MsgId:      msgID,
				}

				buf, err := json.Marshal(readAckRequest)
				if err != nil {
					logger.Errorf("READ_ACK_REQUEST parse json error")
					continue
				}

				length := len(buf)

				typeBuf := make([]byte, 2)
				lengthBuf := make([]byte, 4)

				binary.BigEndian.PutUint32(lengthBuf, uint32(length))
				binary.BigEndian.PutUint16(typeBuf, READ_ACK_REQUEST)

				buf = append(typeBuf[1:], append(lengthBuf, buf[0:]...)[0:]...)

				conn.Write(buf)

				logger.Infof("Send READ_ACK_REQUEST Message:\n%s", buf)

			case "PULL_OLD_MSG_REQUEST":
				logger.Infof("Send PULL_OLD_MSG_REQUEST")

				remoteId, _ := strconv.ParseUint(block[1], 10, 64)
				msgID, _ := strconv.ParseUint(block[2], 10, 64)
				limit, _ := strconv.ParseUint(block[3], 10, 64)

				request := PullOldMsgRequest{
					Rid:      uint64(time.Now().Nanosecond()),
					UserId:   userId,
					RemoteId: remoteId,
					MaxMsgId: msgID,
					Limit:    limit,
				}

				buf, err := json.Marshal(request)
				if err != nil {
					logger.Errorf("PULL_OLD_MSG_REQUEST parse json error")
					continue
				}

				length := len(buf)

				typeBuf := make([]byte, 2)
				lengthBuf := make([]byte, 4)

				binary.BigEndian.PutUint32(lengthBuf, uint32(length))
				binary.BigEndian.PutUint16(typeBuf, PULL_OLD_MSG_REQUEST)

				buf = append(typeBuf[1:], append(lengthBuf, buf[0:]...)[0:]...)

				conn.Write(buf)

				logger.Infof("Send PULL_OLD_MSG_REQUEST Message:\n%s", buf)

			case "SEND_MSG_REQUEST":
				logger.Infof("Send SEND_MSG_REQUEST")

				to, _ := strconv.ParseUint(block[1], 10, 64)
				msg := block[2]

				request := SendMsgRequest{
					Rid:  uint64(time.Now().Nanosecond()),
					From: userId,
					To:   to,
					Msg: Message{
						From: userId,
						To:   to,
						Body: Content{
							Minetype: "plain/text",
							Text:     msg,
						},
					},
				}

				buf, err := json.Marshal(request)
				if err != nil {
					logger.Errorf("PULL_OLD_MSG_REQUEST parse json error")
					continue
				}

				length := len(buf)

				typeBuf := make([]byte, 2)
				lengthBuf := make([]byte, 4)

				binary.BigEndian.PutUint32(lengthBuf, uint32(length))
				binary.BigEndian.PutUint16(typeBuf, SEND_MSG_REQUEST)

				buf = append(typeBuf[1:], append(lengthBuf, buf[0:]...)[0:]...)

				conn.Write(buf)

				logger.Infof("Send SEND_MSG_REQUEST Message:\n%s", buf)

			default:
				logger.Errorf("Unknown request command. %s", command)
			}

		}

		waitGroup.Done()
	}()

	waitGroup.Add(1)

	go func() {
		for {
			buf := make([]byte, 1024*1024)

			_, err := conn.Read(buf)

			if err != nil {
				logger.Errorf("Read info error. Error: %s", err)
				break
			}

			logger.Infof("Receiced Information %s", buf)
		}

		waitGroup.Done()
	}()

	waitGroup.Wait()
}
