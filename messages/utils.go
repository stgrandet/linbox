package messages

import (
	"encoding/binary"
	"encoding/json"
	logger "github.com/cihub/seelog"
	"time"
)

func BuildSyncUnreadRequestBuf(useRId uint64, sessionKey string) (buf []byte, err error) {
	message := SyncUnreadRequest{
		RId:        uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		FromId:     useRId,
		SessionKey: sessionKey,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SyncUnreadRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SYNC_UNREAD_REQUEST, buf), nil
}

func BuildSyncUnreadResponseBuf(RId uint64, useRId uint64, unreadMsgs []UnreadMsg, errCode uint32, errMsg string) (buf []byte, err error) {
	message := SyncUnreadResponse{
		RId:       RId,
		FromId:    useRId,
		Unreads:   unreadMsgs,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SyncUnreadResponse to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SYNC_UNREAD_RESPONSE, buf), nil
}

func BuildReadAckRequestBuf(useRId uint64, sessionKey string, msgId uint64) (buf []byte, err error) {
	message := ReadAckRequest{
		RId:        uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		FromId:     useRId,
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

func BuildReadAckResponseBuf(RId, useRId uint64, sessionKey string, errCode uint32, errMsg string) (buf []byte, err error) {
	message := ReadAckResponse{
		RId:        RId,
		FromId:     useRId,
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

func BuildPullOldMsgRequestBuf(useRId uint64, sessionKey string, maxMsgId, limit uint64) (buf []byte, err error) {
	message := PullOldMsgRequest{
		RId:        uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		FromId:     useRId,
		SessionKey: sessionKey,
		MaxMsgId:   maxMsgId,
		Limit:      limit,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse PullOldMsgRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(PULL_OLD_MSG_REQUEST, buf), nil
}

func BuildPullOldMsgResponseBuf(RId, useRId uint64, msgs []Message, errCode uint32, errMsg string) (buf []byte, err error) {
	message := PullOldMsgResponse{
		RId:       RId,
		FromId:    useRId,
		Msgs:      msgs,
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
		RId:    uint64(time.Now().UnixNano()) / uint64(time.Millisecond),
		FromId: from,
		ToId:   to,
		Msg:    msg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse SendMsgRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(SEND_MSG_REQUEST, buf), nil
}

func BuildSendMsgResponseBuf(RId, from, to, msgId, sendTime uint64, errCode uint32, errMsg string) (buf []byte, err error) {
	message := SendMsgResponse{
		RId:       RId,
		FromId:    from,
		ToId:      to,
		MaxMsgId:  msgId,
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

func BuildAuthRequestBuf(useRId uint64, authMsg map[string]string) (buf []byte, err error) {
	message := AuthRequest{
		FromId:      useRId,
		AuthMessage: authMsg,
	}

	buf, err = json.Marshal(message)
	if err != nil {
		logger.Errorf("Parse AuthRequest to json error. Object: %+v. Error: %s", message, err)
		return buf, err
	}

	return buildBuf(AUTH_REQUEST, buf), nil
}

func BuildAuthResponseBuf(useRId uint64, errCode uint32, errMsg string) (buf []byte, err error) {
	message := AuthResponse{
		FromId:    useRId,
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

func BuildNewMessageBuf(useRId uint64) (buf []byte, err error) {
	message := NewMessage{
		FromId: useRId,
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
