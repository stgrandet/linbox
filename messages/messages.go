package messages

/*
	定义了所有的接口所使用的所有 message 类型
*/

const (
	// 客户端同步未读信息
	SYNC_UNREAD_REQUEST uint16 = iota
	SYNC_UNREAD_RESPONSE

	// 客户端确认未读信息已被读取
	READ_ACK_REQUEST
	READ_ACK_RESPONSE

	// 客户端以反序分页拉取信息
	PULL_OLD_MSG_REQUEST
	PULL_OLD_MSG_RESPONSE

	// 客户端发送信息
	SEND_MSG_REQUEST
	SEND_MSG_RESPONSE

	// 客户端鉴权，同时用于连接建立后试探网络环境
	// 鉴权信息必须是网络联通后客户端发送的第一个信息，否则连接被关闭
	AUTH_REQUEST
	AUTH_RESPONSE

	NEW_MSG_INFO
)

const (
	TOPIC_SYNC_UNREAD_REQUEST  = "im_sync_unread_request"
	TOPIC_READ_ACK_REQUEST     = "im_read_ack_request"
	TOPIC_PULL_OLD_MSG_REQUEST = "im_pull_old_msg_request"
	TOPIC_SEND_MSG_REQUEST     = "im_send_msg_request"
)

const (
	ERROR_WRONG_REQUEST_BODY_FORMAT uint32 = iota
	ERROR_ILLEGAL_SESSION_KEY
	ERROR_EMPTY_INBOX
	ERROR_MSG_NOT_EXIST

	ERROR_AUTH_FAIL

	SUCCESS_RESPONSE uint32 = 200
)

type SyncUnreadRequest struct {
	RId        uint64 `json:"rId"`        // 客户端请求编号，用于在 response 中确认
	FromId     uint64 `json:"fromId"`     // 用户 id
	SessionKey string `json:"sessionKey"` // 要请求的 session key，如果为 nil，则代表请求所有 sessioin 的未读数据
}

type SyncUnreadResponse struct {
	RId       uint64      `json:"rId"`       // 对应客户端所发送请求的编号
	FromId    uint64      `json:"fromId"`    // 用户 id
	Unreads   []UnreadMsg `json:"unreads"`   // 未读信息内容
	ErrorCode uint32      `json:"errorCode"` // 状态吗, 200 表示成功
	ErrorMsg  string      `json:"errorMsg"`  // 具体错误信息
}

type UnreadMsg struct {
	SessionKey string  `json:"sessionKey"` // 未读信息对应的 session
	MsgId      uint64  `json:"msgId"`      //信息编号
	Count      uint64  `json:"count"`      // session 对应的未读信息数
	Msg        Message `json:"msg"`        // 信息内容
}

type ReadAckRequest struct {
	RId        uint64 `json:"rId"`        // 客户端请求编号
	FromId     uint64 `json:"fromId"`     // 用户 id
	SessionKey string `json:"sessionKey"` // 对应的对话 session
	MsgId      uint64 `json:"msgId"`      // 信息编号
}

type ReadAckResponse struct {
	RId        uint64 `json:"rId"`        // 客户端请求编号
	FromId     uint64 `json:"fromId"`     // 用户 id
	SessionKey string `json:"sessionKey"` // 对应的对话 session
	ErrorCode  uint32 `json:"errorCode"`  // 状态吗, 200 表示成功
	ErrorMsg   string `json:"errorMsg"`   // 具体错误信息
}

type PullOldMsgRequest struct {
	RId      uint64 `json:"rId"`      // 客户端请求编号
	FromId   uint64 `json:"fromId"`   // 用户 id
	RemoteId uint64 `json:"remoteId"` //通信对方 id
	MaxMsgId uint64 `json:"maxMsgId"` // 分页起始 id
	Limit    uint64 `json:"limit"`    //分业内数量
}

type PullOldMsgResponse struct {
	RId       uint64    `json:"rId"`       // 客户端请求编号
	FromId    uint64    `json:"fromId"`    // 用户 id
	Msgs      []Message `json:"msgs"`      // 消息内容
	ErrorCode uint32    `json:"errorCode"` // 状态吗, 200 表示成功
	ErrorMsg  string    `json:"errorMsg"`  // 具体错误信息
}

type Message struct {
	FromId     uint64  `json:"fromId"`     // 用户 id
	ToId       uint64  `json:"toId"`       //通信对方 id
	FromUserId uint64  `json:"fromUserId"` // 信息发送方的真实 userId
	MsgId      uint64  `json:"msgId"`      //服务器端产生的唯一 msg id
	Body       Content `json:"body"`       // 消息体
	SendTime   uint64  `json:"sendTime"`   // 服务器端接收到消息的时间戳
}

type Content struct {
	MineType string `json:"mineType"` // 消息内容
	Text     string `json:"text"`     // 文本内容
}

type SendMsgRequest struct {
	RId    uint64  `json:"rId"`    // 客户端请求编号
	FromId uint64  `json:"fromId"` // 用户 id
	ToId   uint64  `json:"toId"`   //通信对方 id
	Msg    Message `json:"msg"`    //消息
}

type SendMsgResponse struct {
	RId       uint64 `json:"rId"`       // 客户端请求编号
	FromId    uint64 `json:"fromId"`    // 用户 id
	ToId      uint64 `json:"toId"`      //通信对方 id
	MaxMsgId  uint64 `json:"maxMsgId"`  //服务器端产生的唯一 msg id, 如果有多个消息，则回传最大 id
	SendTime  uint64 `json:"sendTime"`  // 服务器端接收到消息的时间戳
	ErrorCode uint32 `json:"errorCode"` // 状态吗, 200 表示成功
	ErrorMsg  string `json:"errorMsg"`  // 具体错误信息
}

type AuthRequest struct {
	FromId      uint64            `json:"fromId"`      // user id
	AuthMessage map[string]string `json:"authMessage"` // 目前为空
}

type AuthResponse struct {
	FromId    uint64 `json:"fromId"`
	ErrorCode uint32 `json:"errorCode"` // 状态吗, 200 表示成功
	ErrorMsg  string `json:"errorMsg"`  // 具体错误信息
}

type NewMessage struct {
	FromId uint64 `json:"fromId"` // 信息发送方的 FromId
}
