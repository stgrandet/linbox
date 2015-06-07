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
)

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

func sendEmptySyncUnreadResponse(rid, userId uint64) {
	response, err := BuildSyncUnreadResponseBuf(rid, userId, make([]UnreadMsg, 0), SUCCESS_RESPONSE, "")

	if err != nil {
		logger.Errorf("Build Empty SyncUnreadResponse error. Rid: %d, userId: %d, Error: %s", rid, userId, err)
		return
	}

	sendMessageToUser(userId, response)
}

func sendSyncUnreadResponse(rid, userId uint64, msgs ...UnreadMsg) {
	response, err := BuildSyncUnreadResponseBuf(rid, userId, msgs, SUCCESS_RESPONSE, "")

	if err != nil {
		logger.Errorf("Build Success SyncUnreadResponse error. Rid: %d, userId: %d, Error: %s", rid, userId, err)
		return
	}

	sendMessageToUser(userId, response)
}

func sendErrorSyncUnreadResponse(rid, userId uint64, errCode uint32, errString string) {
	response, err := BuildSyncUnreadResponseBuf(rid, userId, make([]UnreadMsg, 0), errCode, errString)

	if err != nil {
		logger.Errorf("Build Fail SyncUnreadResponse error. Rid: %d, userId: %d, Error: %s", rid, userId, err)
		return
	}

	sendMessageToUser(userId, response)
}

func sendErrorReadAckResponse(rid, userId uint64, sessionKey string, errCode uint32, errString string) {
	response, err := BuildReadAckResponseBuf(rid, userId, sessionKey, errCode, errString)

	if err != nil {
		logger.Errorf("Build ReadAckResponse Error. rid: %d, userid: %d, sessionKey: string, errCode: %d, errString: %s. Error: %s", rid, userId, sessionKey, errCode, errString, err)
		return
	}

	sendMessageToUser(userId, response)
}

func sendReadAckResponse(rid, userId uint64, sessionKey string) {
	response, err := BuildReadAckResponseBuf(rid, userId, sessionKey, SUCCESS_RESPONSE, "")

	if err != nil {
		logger.Errorf("Build ReadAckResponse Error. rid: %d, userid: %d, sessionKey: string. Error: %s", rid, userId, sessionKey, err)
		return
	}

	sendMessageToUser(userId, response)
}

func createSessionKey(id1, id2 uint64) string {
	if id1 < id2 {
		return strconv.FormatUint(id1, 10) + strconv.FormatUint(id2, 10)
	} else {
		return strconv.FormatUint(id2, 10) + strconv.FormatUint(id1, 10)
	}
}

func sendSuccessPullOldMsgResponse(rid, userId uint64, msgs []Message) {
	response, err := BuildPullOldMsgResponseBuf(rid, userId, msgs, SUCCESS_RESPONSE, "")

	if err != nil {
		logger.Errorf("Build PullOldMsgResponse error when sendSuccessPullOldMsgResponse. Rid: %d. UserId: %d. Messages: %+v. Error: %s", rid, userId, msgs, err)
		return
	}

	sendMessageToUser(userId, response)
}

func sendFailPullOldMsgResponse(rid, userId uint64, msgs []Message, errCode uint32, errStr string) {
	response, err := BuildPullOldMsgResponseBuf(rid, userId, msgs, errCode, errStr)

	if err != nil {
		logger.Errorf("Build PullOldMsgResponse error when sendSuccessPullOldMsgResponse. Rid: %d. UserId: %d. Messages: %+v. Error: %s", rid, userId, msgs, err)
		return
	}

	sendMessageToUser(userId, response)
}

func startServer() {
	service := &testConnService{}

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

	conn.SetKeepAlive(true)

	authMsg, err := BuildAuthRequestBuf(userId, nil)
	if err != nil {
		logger.Error("Build auth message error. Error: %s", err)
		return
	}

	conn.Write(authMsg)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			consoleBuf := make([]byte, 200)

			length, err := reader.Read(consoleBuf)
			if err != nil {
				logger.Infof("Read Console error: %s\n", err)
				break
			}

			consoleBuf = consoleBuf[0:length]

			block := strings.Split(string(consoleBuf), " ")
			command := strings.TrimSpace(block[0])

			switch command {
			case "SYNC_UNREAD_REQUEST":
				sendSyncUnreadRequest(conn, userId, "")
			case "READ_ACK_REQUEST":
				sessionKey := block[1]
				msgId, _ := strconv.ParseUint(block[2], 10, 64)

				sendReadAckRequest(conn, userId, sessionKey, msgId)
			case "PULL_OLD_MSG_REQUEST":
				remoteId, _ := strconv.ParseUint(block[1], 10, 64)
				msgId, _ := strconv.ParseUint(block[2], 10, 64)
				limit, _ := strconv.ParseUint(block[3], 10, 64)

				sendPullOldMsgRequest(conn, userId, remoteId, msgId, limit)
			case "SEND_MSG_REQUEST":
				to, _ := strconv.ParseUint(block[1], 10, 64)
				msg := block[2]

				sendSendMsgRequest(conn, userId, to, msg)
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

			length, err := conn.Read(buf)

			if err != nil {
				logger.Errorf("Read info error. Error: %s", err)
				break
			}

			buf = buf[0:length]

			logger.Infof("Receiced Information %s", buf)

			msgTypeByte := append(make([]byte, 1), buf[0])
			msgType := uint16(binary.BigEndian.Uint16(msgTypeByte))

			switch msgType {
			case NEW_MSG_INFO:
				newMsg := &NewMessage{}
				msgBuf := buf[5:]
				err := json.Unmarshal(msgBuf, newMsg)

				if err != nil {
					logger.Errorf("The NewMessage can not be converted from json. Json: %s. Error: %s.", msgBuf, err)
					continue
				}

				remoteId := newMsg.UserId
				sessionKey := createSessionKey(userId, remoteId)

				sendSyncUnreadRequest(conn, userId, sessionKey)

			case SYNC_UNREAD_RESPONSE:
				response := &SyncUnreadResponse{}
				msgBuf := buf[5:]
				err := json.Unmarshal(msgBuf, response)

				if err != nil {
					logger.Errorf("The SyncUnreadResponse can not be converted from json. Json: %s. Error: %s.", msgBuf, err)
					continue
				}

				unreadMsgs := response.Unread
				for _, unread := range unreadMsgs {
					remoteId, err := getRemoteId(unread.SessionKey, userId)
					if err != nil {
						logger.Errorf("Can not find remote id from unread message. Message: %+v. userId: %s.", unread, userId)
						continue
					}

					sendReadAckRequest(conn, userId, unread.SessionKey, unread.MsgId)
					sendPullOldMsgRequest(conn, userId, remoteId, unread.MsgId, unread.Count)
				}

			default:
				continue
			}
		}

		waitGroup.Done()
	}()

	waitGroup.Wait()
}

func sendSyncUnreadRequest(conn *net.TCPConn, userId uint64, sessionKey string) {
	syncUnreadRequestBuf, err := BuildSyncUnreadRequestBuf(userId, "")
	if err != nil {
		logger.Errorf("Build SYNC_UNREAD_REQUEST error. Error: %s", err)
		return
	}
	conn.Write(syncUnreadRequestBuf)

	logger.Infof("Send SYNC_UNREAD_REQUEST Message:\n%s", syncUnreadRequestBuf)
}

func sendReadAckRequest(conn *net.TCPConn, userId uint64, sessionKey string, msgId uint64) {
	buf, err := BuildReadAckRequestBuf(userId, sessionKey, msgId)
	if err != nil {
		logger.Errorf("Build READ_ACK_REQUEST error. Error: %s", err)
		return
	}

	conn.Write(buf)

	logger.Infof("Send READ_ACK_REQUEST Message:\n%s", buf)
}

func sendPullOldMsgRequest(conn *net.TCPConn, from, to, msgId, limit uint64) {
	buf, err := BuildPullOldMsgRequestBuf(from, to, msgId, limit)
	if err != nil {
		logger.Errorf("Build PULL_OLD_MSG_REQUEST error")
		return
	}

	conn.Write(buf)

	logger.Infof("Send PULL_OLD_MSG_REQUEST Message:\n%s", buf)
}

func sendSendMsgRequest(conn *net.TCPConn, from, to uint64, msg string) {
	message := Message{
		From: from,
		To:   to,
		Body: Content {
			Minetype: "plain/text",
			Text:     msg,
		},
	}

	msgs := make([]Message, 1)
	msgs[0] = message

	buf, err := BuildSendMsgRequestBuf(from, to, msgs)
	if err != nil {
		logger.Errorf("Build PULL_OLD_MSG_REQUEST error. Error: %s", err)
		return
	}

	conn.Write(buf)

	logger.Infof("Send SEND_MSG_REQUEST Message:\n%s", buf)
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
