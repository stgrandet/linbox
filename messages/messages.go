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
	ERROR_WRONG_REQUEST_BODY_FORMAT uint32= iota
	ERROR_ILLEGAL_SESSION_KEY
	ERROR_EMPTY_INBOX

	ERROR_AUTH_FAIL

	SUCCESS_RESPONSE uint32 = 200
)

type SyncUnreadRequest struct {
	Rid        uint64 // 客户端请求编号，用于在 response 中确认
	UserId     uint64 // 用户 id
	SessionKey string // 要请求的 session key，如果为 nil，则代表请求所有 sessioin 的未读数据
}

type SyncUnreadResponse struct {
	Rid       uint64      // 对应客户端所发送请求的编号
	UserId    uint64      // 用户 id
	Unread    []UnreadMsg // 未读信息内容
	ErrorCode uint32      // 状态吗, 200 表示成功
	ErrorMsg  string      // 具体错误信息
}

type UnreadMsg struct {
	SessionKey string  // 未读信息对应的 session
	MsgId      uint64  //信息编号
	Count      uint64  // session 对应的未读信息数
	Msg        Message // 信息内容
}

type ReadAckRequest struct {
	Rid        uint64 // 客户端请求编号
	UserId     uint64 // 用户 id
	SessionKey string // 对应的对话 session
	MsgId      uint64 // 信息编号
}

type ReadAckResponse struct {
	Rid        uint64 // 客户端请求编号
	UserId     uint64 // 用户 id
	SessionKey string // 对应的对话 session
	ErrorCode  uint32 // 状态吗, 200 表示成功
	ErrorMsg   string // 具体错误信息
}

type PullOldMsgRequest struct {
	Rid      uint64 // 客户端请求编号
	UserId   uint64 // 用户 id
	RemoteId uint64 //通信对方 id
	MaxMsgId uint64 // 分页起始 id
	Limit    uint64 //分业内数量
}

type PullOldMsgResponse struct {
	Rid       uint64  // 客户端请求编号
	UserId    uint64  // 用户 id
	Msg       Message // 消息内容
	ErrorCode uint32  // 状态吗, 200 表示成功
	ErrorMsg  string  // 具体错误信息
}

type Message struct {
	From     uint64  // 用户 id
	To       uint64  //通信对方 id
	MsgId    uint64  //服务器端产生的唯一 msg id
	Body     Content // 消息体
	SentTime uint64  // 服务器端接收到消息的时间戳
}

type Content struct {
	Minetype string // 消息内容
	Text     string // 文本内容
	MediaId  uint64 // 音频或视频文件 id
}

type SendMsgRequest struct {
	Rid  uint64  // 客户端请求编号
	From uint64  // 用户 id
	To   uint64  //通信对方 id
	Msg  Message //消息体
}

type SendMsgResponse struct {
	Rid       uint64 // 客户端请求编号
	From      uint64 // 用户 id
	To        uint64 //通信对方 id
	MsgId     uint64 //服务器端产生的唯一 msg id
	SendTime  uint64 // 服务器端接收到消息的时间戳
	ErrorCode uint32 // 状态吗, 200 表示成功
	ErrorMsg  string // 具体错误信息
}

type AuthRequest struct {
	UserId uint64  // user id
	AuthMessage map[string]string // 目前为空
}

type AuthResponse struct {
	UserId uint64
	ErrorCode uint32 // 状态吗, 200 表示成功
	ErrorMsg  string // 具体错误信息
}

type NewMessage struct {
	UserId uint64 // 信息发送方的 userid
}
