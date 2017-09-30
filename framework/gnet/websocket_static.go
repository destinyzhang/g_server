package gnet

import (
	"errors"
)

const (
	_buffCapHead = 14
	_buffCap     = 2048

	_wsHkHead       = "Head"
	_wsHkProtocol   = "Sec-WebSocket-Protocol"
	_wsHkOrigin     = "Origin"
	_wsHkConnection = "Connection"
	_wsHkHost       = "Host"
	_wsHkUpgrade    = "Upgrade"
	_wsHkVersion    = "Sec-WebSocket-Version"
	_wsHkKey        = "Sec-WebSocket-Key"
	_wsHkAccept     = "Sec-WebSocket-Accept"

	_wsOpcodeCon   = byte(0x0)
	_wsOpcodeTxt   = byte(0x1)
	_wsOpcodeBit   = byte(0x2)
	_wsOpcodeClose = byte(0x8)
	_wsOpcodePing  = byte(0x9)
	_wsOpcodePong  = byte(0xA)

	WsStateClosed     = 0
	WsStateCloseing   = 1
	WsStateConnecting = 2
	WsStateConnected  = 3

	WsServerStateClosed   = 0
	WsServerStateCloseing = 1
	WsServerListenning    = 2
)

var (
	_wsMagicKey = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	_wsCrlf     = []byte("\r\n")

	ErrRSV123         = errors.New("Err RSV123")
	ErrMsgSizeInvalid = errors.New("Err MsgSizeInvalid")
	ErrInvalidOpcode  = errors.New("Err InvalidOpcode")
	ErrHandshakeEmpty = errors.New("Err ErrHandshakeEmpty")
	ErrHandshake      = errors.New("Err Handshake")
)

//webSocketMsg 发送消息使用
type webSocketMsg struct {
	buff   []byte
	opcode byte
}

//NewWebSocketServer 生成一个服务器
func NewWebSocketServer(shost string, maxmsgsize uint32) *WebSocketServer {
	return &WebSocketServer{BaseServer: BaseServer{host: shost}, wsmaxmsgsize: maxmsgsize}
}

//NewWebSocketServerSimple 生成一个服务器
func NewWebSocketServerSimple(shost string, maxmsgsize uint32) *WebSocketServerSimple {
	return &WebSocketServerSimple{WebSocketServer: WebSocketServer{BaseServer: BaseServer{host: shost}, wsmaxmsgsize: maxmsgsize}}
}

//NewWebSocketClient 生成一个客户端
func NewWebSocketClient(curl string, maxmsgsize uint32) *WebSocketClient {
	return &WebSocketClient{WebSocket: WebSocket{wsmaxmsgsize: maxmsgsize}, hosturl: curl, network: "tcp"}
}

//NewWebSocketClient 生成一个客户端
func NewWebSocketIpv6Client(curl string, maxmsgsize uint32) *WebSocketClient {
	return &WebSocketClient{WebSocket: WebSocket{wsmaxmsgsize: maxmsgsize}, hosturl: curl, network: "tcp6"}
}
