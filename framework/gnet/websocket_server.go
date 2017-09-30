package gnet

import (
	"bufio"
	"net"
)

//WebSocketServer 服务器
type WebSocketServer struct {
	BaseServer
	wsmaxmsgsize uint32
	watcher      IServerWatcher
}

func (server *WebSocketServer) TypeName() string {
	return "websocketserver"
}

//SetMaxMsgSize 设置接受最大包大小
func (server *WebSocketServer) SetMaxMsgSize(size uint32) {
	server.wsmaxmsgsize = size
}

//GetMaxMsgSize 返回接受最大包大小
func (server *WebSocketServer) GetMaxMsgSize() uint32 {
	return server.wsmaxmsgsize
}

//SetWatcher
func (server *WebSocketServer) SetWatcher(watcher IServerWatcher) {
	server.watcher = watcher
}

//SetWatcher
func (server *WebSocketServer) GetWatcher() IServerWatcher {
	return server.watcher
}

//Start 开启
func (server *WebSocketServer) Start() bool {
	server.fnewConn = server.newWebSocket
	server.network = "tcp"
	return server.BaseServer.Start()
}

func (server *WebSocketServer) newWebSocket(conn net.Conn) {
	ws := &WebSocket{
		conn:         conn,
		rw:           bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		needmask:     false,
		state:        WsStateConnecting,
		connid:       server.genConnid(),
		wsmaxmsgsize: server.wsmaxmsgsize}
	if server.watcher != nil {
		server.watcher.OnSocketAccept(ws)
	}
	ws.Start()
}
