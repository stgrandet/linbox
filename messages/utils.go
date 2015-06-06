package messages

import (
	"encoding/binary"
	"encoding/json"
	logger "github.com/cihub/seelog"
	"time"
)

func BuildSyncUnreadRequestBuf(userId uint64, sessionKey string) (buf []byte, err error) {
	message := SyncUnreadRequest{
		Rid:        uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		UserId:     userId,
		SessionKey: sessionKey,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SyncUnreadRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SYNC_UNREAD_REQUEST, buf)
}

func BuildSynbcUnreadResponseBuf(rid uint64, userId uint64, unreadMsgs []UnreadMsg, errCode uint32, errMsg string) (buf []byte, err error) {
	message := SyncUnreadResponse{
		Rid:       rid,
		UserId:    userId,
		Unread:    unreadMsgs,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SyncUnreadResponse to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SYNC_UNREAD_RESPONSE, buf)
}

func BuildReadAckRequestBuf(userId uint64, sessionKey string, msgId uint64) (buf []byte, err error) {
	message := ReadAckRequest{
		Rid:        uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		UserId:     userId,
		SessionKey: sessionKey,
		MsgId:      msgId,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse ReadAckRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(READ_ACK_REQUEST, buf), nil
}

func BuildReadAckResponseBuf(rid, userId uint64, sessionKey string, errCode uint32, errMsg string) (buf []byte, err error) {
	message := ReadAckResponse{
		Rid:        rid,
		UserId:     userId,
		SessionKey: sessionKey,
		ErrorCode:  errCode,
		ErrorMsg:   errMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse ReadAckResponse to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(READ_ACK_RESPONSE, buf), nil
}

func BuildPullOldMsgRequestBuf(userId, remoteId, maxMsgId, limit uint64) (buf []byte, err error) {
	message := PullOldMsgRequest{
		Rid:      uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		UserId:   userId,
		RemoteId: remoteId,
		MaxMsgId: maxMsgId,
		Limit:    limit,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse PullOldMsgRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(READ_ACK_RESPONSE, buf), nil
}

func BuildPullOldMsgResponseBuf(rid, userId uint64, msg Message, errCode uint32, errMsg string) (buf []byte, err error) {
	message := PullOldMsgResponse{
		Rid:       rid,
		UserId:    userId,
		Msg:       msg,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse PullOldMsgResponse to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(PULL_OLD_MSG_RESPONSE, buf), nil
}

func BuildSendMsgRequestBuf(from, to uint64, msg Message) (buf []byte, err error) {
	message := SendMsgRequest{
		Rid:  uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		From: from,
		To:   to,
		Msg:  msg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SendMsgRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SEND_MSG_REQUEST, buf), nil
}

func BuildSendMsgResponseBuf(rid, from, to, msgId, sendTime uint64, errCode uint32, errMsg string) (buf []byte, err error) {
	message := SendMsgResponse{
		Rid:       rid,
		From:      from,
		To:        to,
		MsgId:     msgId,
		SendTime:  sendTime,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SendMsgResponse to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SEND_MSG_RESPONSE, buf), nil
}

func BuildAuthRequestBuf(userId uint64, authMsg map[string]string) (buf []byte, err error) {
	message := AuthRequest{
		UserId:      userId,
		AuthMessage: authMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse AuthRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(AUTH_REQUEST, buf), nil
}

func BuildAuthResponseBuf(userId uint64, errCode uint32, errMsg string) (buf []byte, err error) {
	message := AuthResponse{
		UserId:    userId,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse AuthResponse to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(AUTH_RESPONSE, buf), nil
}

func BuildNewMessageBuf(userId uint64) (buf []byte, err error) {
	message := NewMessage{
		UserId: userId,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse NewMessage to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(NEW_MSG_INFO, buf), nil
}

func buildBuf(msgType uint16, buf []byte) []byte {
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(buf)))

	typeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBuf, msgType)

	buf = append(typeBuf[1:], append(lengthBuf, buf[0:]...)[0:]...)

	return buf
}
