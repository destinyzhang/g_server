package gnet

import (
	"sync"
)

//WebSocketServerSimple 服务器
type WebSocketServerSimple struct {
	WebSocketServer
	sync.RWMutex
	mapws map[uint64]*WebSocket

	clientOpen  func(uint64)
	clientClose func(uint64)
	clientMsg   func(uint64, []byte)
}

//OnClientOpen 设置事件回调，这个是多个协程同时访问
func (server *WebSocketServerSimple) OnClientOpen(evt func(uint64)) {
	server.clientOpen = evt
}

//OnClientClose 设置事件回调，这个是多个协程同时访问
func (server *WebSocketServerSimple) OnClientClose(evt func(uint64)) {
	server.clientClose = evt
}

//OnClientMsg 设置事件回调，这个是多个协程同时访问
func (server *WebSocketServerSimple) OnClientMsg(evt func(uint64, []byte)) {
	server.clientMsg = evt
}

//SendText 发送文本
func (server *WebSocketServerSimple) SendText(id uint64, msg string) {
	if ws, ok := server.getClient(id); ok {
		ws.SendText(msg)
	}
}

//SendBit 发送二进制
func (server *WebSocketServerSimple) SendBit(id uint64, msg []byte) {
	if ws, ok := server.getClient(id); ok {
		ws.SendBit(msg)
	}
}

//KickAll 踢掉连接
func (server *WebSocketServerSimple) KickAll() {
	server.RLock()
	defer server.RUnlock()
	for _, ws := range server.mapws {
		ws.Close()
	}
}

//Kick 踢掉指定连接
func (server *WebSocketServerSimple) Kick(id uint64) {
	if ws, ok := server.getClient(id); ok {
		ws.Close()
	}
}

//BroadcastBit 广播二进制
func (server *WebSocketServerSimple) BroadcastBit(msg []byte) {
	server.RLock()
	defer server.RUnlock()
	for _, ws := range server.mapws {
		ws.SendBit(msg)
	}
}

//BroadcastText 广播文本
func (server *WebSocketServerSimple) BroadcastText(msg string) {
	server.RLock()
	defer server.RUnlock()
	for _, ws := range server.mapws {
		ws.SendText(msg)
	}
}

//Stop 关闭
func (server *WebSocketServerSimple) Stop() bool {
	if reslut := server.WebSocketServer.Stop(); reslut {
		server.mapws = nil
		return true
	}
	return false
}

//Start 开启
func (server *WebSocketServerSimple) Start() bool {
	if reslut := server.WebSocketServer.Start(); reslut {
		server.mapws = make(map[uint64]*WebSocket)
		server.SetWatcher(server)
		return true
	}
	return false
}

//ClientCount 连接个数
func (server *WebSocketServerSimple) ClientCount() int {
	if server.state != WsServerListenning {
		return 0
	}
	server.RLock()
	defer server.RUnlock()
	return len(server.mapws)
}

func (server *WebSocketServerSimple) addClientToMap(ws *WebSocket) {
	server.Lock()
	defer server.Unlock()
	server.mapws[ws.ID()] = ws
}

func (server *WebSocketServerSimple) removeClientFromMap(ws *WebSocket) {
	server.Lock()
	defer server.Unlock()
	delete(server.mapws, ws.ID())
}

func (server *WebSocketServerSimple) getClient(id uint64) (*WebSocket, bool) {
	server.RLock()
	defer server.RUnlock()
	if ws, ok := server.mapws[id]; ok {
		return ws, true
	}
	return nil, false
}

func (server *WebSocketServerSimple) OnSocketAccept(ws ISocket) {
	ws.SetWatcher(server)
}

func (server *WebSocketServerSimple) OnSocketOpen(ws ISocket) {
	server.addClientToMap(ws.(*WebSocket))
	if server.clientOpen != nil {
		server.clientOpen(ws.ID())
	}
}

func (server *WebSocketServerSimple) OnSocketClose(ws ISocket) {
	server.removeClientFromMap(ws.(*WebSocket))
	if server.clientClose != nil {
		server.clientClose(ws.ID())
	}
}

func (server *WebSocketServerSimple) OnSocketMessage(ws ISocket, buff []byte) {
	if server.clientMsg != nil {
		server.clientMsg(ws.ID(), buff)
	}
}
