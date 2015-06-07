package main

import (
	"encoding/json"
	logger "github.com/cihub/seelog"
	"sync"
	"time"

	"linbox/connector"
	. "linbox/messages"
)

var (
	messageCenter = make(map[string]map[uint64]Message)   // 消息存储中心
	inBox         = make(map[uint64]map[string]UnreadMsg) // 未读信息收件箱
	messageBox    = make(map[uint64][][]byte)             // 发送信息箱
	msgIdRecord   = make(map[string]uint64)               // msg id 存储器

	messageCenterRwMutex sync.RWMutex // 消息中心的读写所
	inBoxRwMutex         sync.RWMutex // 发件箱读写锁
	messageBoxRWMutex    sync.RWMutex // 消息读写锁
	msgIdMutex           sync.Mutex   // Id生成器的锁
)

type testConnService struct {
	connector.MqService
}

func (*testConnService) InitService() {

}

func (s *testConnService) HandleReceivingMsg(msgType uint16, msg []byte) {
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

func (*testConnService) handleSyncUnreadRequest(msg []byte) {
	logger.Infof("Received SYNC_UNREAD_REQUEST message %s", msg)

	syncUnreadRequest := &SyncUnreadRequest{}
	err := json.Unmarshal(msg, syncUnreadRequest)
	if err != nil {
		logger.Errorf("Can not parse SyncUnreadRequest from json. Error: %s", err)
		return
	}

	rid := syncUnreadRequest.RId
	userId := syncUnreadRequest.FromId
	sessionKey := syncUnreadRequest.SessionKey

	if sessionKey != "" {
		defer inBoxRwMutex.RUnlock()
		inBoxRwMutex.RLock()

		allMessages, ok := inBox[userId]
		if !ok || len(allMessages) < 0 {
			sendEmptySyncUnreadResponse(rid, userId)
			return
		}

		unReadMsg, ok := allMessages[sessionKey]
		if !ok {
			sendEmptySyncUnreadResponse(rid, userId)
			return
		}

		sendSyncUnreadResponse(rid, userId, unReadMsg)
		return
	} else {

		defer inBoxRwMutex.RUnlock()
		inBoxRwMutex.RLock()

		// 拉取所有的未读信息
		allMessages, ok := inBox[userId]
		if !ok || len(allMessages) < 0 {
			sendEmptySyncUnreadResponse(rid, userId)
			return
		} else {
			sendSyncUnreadResponse(rid, userId, allMessages[sessionKey])
			return
		}
	}
}

func (*testConnService) handleReadAckRequest(msg []byte) {
	logger.Infof("Received READ_ACK_REQUEST message %s", msg)

	readAckRequest := &ReadAckRequest{}
	err := json.Unmarshal(msg, readAckRequest)
	if err != nil {
		logger.Errorf("Can not parse ReadAckRequest from json. Error: %s", err)
		return
	}

	rid := readAckRequest.RId
	userId := readAckRequest.FromId
	sessionKey := readAckRequest.SessionKey
	msgId := readAckRequest.MsgId

	defer inBoxRwMutex.Unlock()
	inBoxRwMutex.Lock()

	allMessages, ok := inBox[userId]
	if !ok || len(allMessages) < 0 {
		sendErrorReadAckResponse(rid, userId, sessionKey, ERROR_EMPTY_INBOX, "No messages found for user")
		return
	}

	unReadMsg, ok := allMessages[sessionKey]
	if !ok {
		sendErrorReadAckResponse(rid, userId, sessionKey, ERROR_EMPTY_INBOX, "No messages found for session key")
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

	sendReadAckResponse(rid, userId, sessionKey)
}

func (*testConnService) handlePullOldMsgRequest(msg []byte) {
	logger.Infof("Received PULL_OLD_MSG_REQUEST message %s", msg)

	pullOldRequest := &PullOldMsgRequest{}
	err := json.Unmarshal(msg, pullOldRequest)
	if err != nil {
		logger.Errorf("Can not parse PullOldMsgRequest from json. Error: %s", err)
		return
	}

	rid := pullOldRequest.RId
	userId := pullOldRequest.FromId
	remoteId := pullOldRequest.RemoteId
	maxMsgId := pullOldRequest.MaxMsgId
	limit := pullOldRequest.Limit
	sessionKey := createSessionKey(userId, remoteId)

	msgs := getPersistMsg(sessionKey, maxMsgId, limit)

	if len(msgs) <= 0 {
		sendFailPullOldMsgResponse(rid, userId, msgs, ERROR_MSG_NOT_EXIST, "Can not find corresponding messages")
		return
	}

	sendSuccessPullOldMsgResponse(rid, userId, msgs)
}

func (*testConnService) handleSendMsgRequest(request []byte) {
	logger.Infof("Received SEND_MSG_REQUEST message %s", request)

	sendMsgRequest := &SendMsgRequest{}
	err := json.Unmarshal(request, sendMsgRequest)
	if err != nil {
		logger.Errorf("Can not parse SendMsgRequest from json. Error: %s", err)
		return
	}

	rid := sendMsgRequest.RId
	from := sendMsgRequest.FromId
	to := sendMsgRequest.ToId
	msg := sendMsgRequest.Msg
	sessionKey := createSessionKey(from, to)

	msg = genMsgId(sessionKey, msg)[0]
	persistMsg(sessionKey, msg)

	shouldPushNew := addToInBox(sessionKey, to, msg)

	if shouldPushNew {
		pushNew(from, to)
	}

	sendSendMsgResponse(from, to, rid, msg.MsgId, sessionKey)
}

func (s *testConnService) HandleSendingMsg(userid uint64, listenChannel chan<- []byte, quit <-chan bool) {
	go func() {
		defer close(listenChannel)

		loop:
		for {
			timer := time.NewTimer(time.Millisecond * time.Duration(500))

			select {
			case <-timer.C:
				messageBoxRWMutex.Lock()

				if messages, ok := messageBox[userid]; ok {
					for _, message := range messages {
						listenChannel <- message
						logger.Infof("Send message: %s", message)
					}

					delete(messageBox, userid)
				}

				messageBoxRWMutex.Unlock()
			case <-quit:
				logger.Info("Read quit info. Close connection")
				break loop
			}
		}
	}()
}

func genMsgId(sessionKey string, msgs ...Message) []Message {
	defer msgIdMutex.Unlock()
	msgIdMutex.Lock()

	maxMsgId, ok := msgIdRecord[sessionKey]
	if !ok {
		maxMsgId = 0
	}

	for i, msg := range msgs {
		maxMsgId++
		msg.MsgId = maxMsgId
		msgs[i] = msg
	}

	msgIdRecord[sessionKey] = maxMsgId

	return msgs
}

func getPersistMsg(sessionKey string, maxMsgID uint64, count uint64) (msgs []Message) {
	msgs = make([]Message, 0)

	defer messageCenterRwMutex.RUnlock()
	messageCenterRwMutex.RLock()

	allMsg, existed := messageCenter[sessionKey]
	if !existed {
		logger.Errorf("Can not find msgs for session %s msgId: %d", sessionKey, maxMsgID)
		return
	}

	for i := uint64(0); i < count; i++ {
		msg, existed := allMsg[maxMsgID-i]
		if !existed {
			logger.Errorf("Can not find message session: %s. msgId: %d", sessionKey, maxMsgID-i)
			continue
		}

		msgs = append(msgs, msg)
	}

	return
}

func persistMsg(sessionKey string, msgs ...Message) {
	defer messageCenterRwMutex.Unlock()
	messageCenterRwMutex.Lock()

	allMsg, existed := messageCenter[sessionKey]
	if !existed {
		allMsg = make(map[uint64]Message, len(msgs))
	}

	for _, msg := range msgs {
		allMsg[msg.MsgId] = msg
	}

	messageCenter[sessionKey] = allMsg
}

func addToInBox(sessionKey string, to uint64, msg Message) (isNew bool) {
	defer inBoxRwMutex.Unlock()
	inBoxRwMutex.Lock()

	isNew = false

	allMessages, ok := inBox[to]
	if !ok {
		allMessages = make(map[string]UnreadMsg)
	}

	unreadMsg, ok := allMessages[sessionKey]
	if !ok {
		unreadMsg = UnreadMsg{
			SessionKey: sessionKey,
			MsgId:      0,
			Count:      0,
		}
	}

	if unreadMsg.MsgId < msg.MsgId {
		unreadMsg.Count += msg.MsgId - unreadMsg.MsgId
		unreadMsg.MsgId = msg.MsgId
		unreadMsg.Msg = msg

		isNew = true

		allMessages[sessionKey] = unreadMsg
		inBox[to] = allMessages
	}

	logger.Infof("Add unreadMsg to inbox %d. msg: %+v", to, unreadMsg)

	return
}

func sendSendMsgResponse(userId, remoteId, rid, maxMsgId uint64, sessionKey string) {
	message, err := BuildSendMsgResponseBuf(rid, userId, remoteId, maxMsgId, uint64(time.Now().Nanosecond())/uint64(time.Millisecond), SUCCESS_RESPONSE, "")
	if err != nil {
		logger.Errorf("Send SendMsgResponse Error. Error: %s", err)
		return
	}

	sendMessageToUser(userId, message)
}

func pushNew(from, to uint64) {
	message, err := BuildNewMessageBuf(from)

	if err != nil {
		logger.Errorf("Build NewMessage buf fail. From: %d. To: %d. Error: %d", from, to, err)
		return
	}

	sendMessageToUser(to, message)
}

func sendMessageToUser(userId uint64, message []byte) {
	defer messageBoxRWMutex.Unlock()
	messageBoxRWMutex.Lock()

	msgList, ok := messageBox[userId]
	if !ok {
		msgList = make([][]byte, 0)
	}

	msgList = append(msgList, message)

	messageBox[userId] = msgList
}
