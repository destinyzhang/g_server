package session

import (
	"g_server/framework/gnet"
)

const (
	sessionEventOpen  = 1
	sessionEventClose = 2
	sessionEventMsg   = 3
)

func NewWsSessionManager(name string, host string, maxmsgsize uint32, maxsession uint32, hook func(gnet.ISocket, []byte) bool) *SessionManager {
	return &SessionManager{server: gnet.NewWebSocketServer(host, maxmsgsize), SessionMsgProxy: SessionMsgProxy{msghanders: make(map[uint32]*msgProxy)}, maxsession: maxsession, name: name, hook: hook}
}

func NewWsSessionClient(name string, curl string, rcontime int32, maxsession uint32) *SessionClient {
	return &SessionClient{BaseSession: BaseSession{ws: gnet.NewWebSocketClient(curl, maxsession)}, SessionMsgProxy: SessionMsgProxy{msghanders: make(map[uint32]*msgProxy)}, rcontime: rcontime, name: name}
}
