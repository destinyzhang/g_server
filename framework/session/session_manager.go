package session

import (
	"g_server/framework/com"
	"g_server/framework/gnet"
	"g_server/framework/msgpack"
	"g_server/framework/protocolbase"
)

type Session struct {
	BaseSession
	SessionMsgQueue
	manager *SessionManager
	skip    bool
}

func (session *Session) OnSocketOpen(ws gnet.ISocket) {
	session.manager.msgrec.Push(&sessionEvent{event: sessionEventOpen, ws: ws})
}

func (session *Session) OnSocketClose(ws gnet.ISocket) {
	session.ws.SetWatcher(nil)
	session.manager.msgrec.Push(&sessionEvent{event: sessionEventClose, ws: ws})
}

func (session *Session) OnSocketMessage(ws gnet.ISocket, msg []byte) {
	if session.manager.hook != nil {
		skip := false
		com.SafeCall(func() {
			skip = session.manager.hook(ws, msg)
		})
		if skip {
			return
		}
		session.msgrec.Push(msg)
	} else {
		session.msgrec.Push(msg)
	}

}

func (session *Session) handleMsg(unpacker protocolbase.IUnpacker) {
	session.copymsg()
	for {
		ibyte := session.msghand.Pop()
		if ibyte == nil {
			break
		}
		if session.skip {
			continue
		}
		unpacker.Attatch(ibyte.([]byte))
		if r, id := unpacker.UnPackUInt32(); r == 0 {
			if msgProxy := session.manager.FindMsgProxy(id); msgProxy != nil {
				com.SafeCall(func() {
					if msg := msgProxy.msgCreate(); msg != nil {
						msgProxy.msgHandler(session, msg, msg.Unpack(unpacker) == 0)
					}
				})
			}
		}
	}
}

func (session *Session) SkipMsg(skip bool) {
	session.skip = skip
}

type SessionManager struct {
	SessionMsgProxy
	SessionMsgQueue
	server     gnet.IServer
	ssmap      map[uint64]*Session
	maxsession uint32
	name       string
	hook       func(gnet.ISocket, []byte) bool
}

func (manager *SessionManager) OnSocketAccept(ws gnet.ISocket) {
	session := &Session{BaseSession: BaseSession{ws: ws}, manager: manager, skip: false}
	session.init()
	ws.SetWatcher(session)
}

func (manager *SessionManager) getSession(id uint64) *Session {
	session, _ := manager.ssmap[id]
	return session
}

func (manager *SessionManager) handleEvent() {
	manager.copymsg()
	for {
		ievent := manager.msghand.Pop()
		if ievent == nil {
			break
		}
		event := ievent.(*sessionEvent)
		switch event.event {
		case sessionEventOpen:
			{
				if manager.maxsession <= uint32(manager.Count()) {
					event.ws.Close()
					continue
				}
				session := event.ws.GetWatcher().(*Session)
				manager.ssmap[event.ws.ID()] = session
				if manager.fsessionOpen != nil {
					manager.fsessionOpen(session)
				}
			}
		case sessionEventClose:
			{
				if session := manager.getSession(event.ws.ID()); session != nil {
					delete(manager.ssmap, event.ws.ID())
					if manager.fSessionClose != nil {
						manager.fSessionClose(session)
					}
				}
			}
		}
	}
}

func (manager *SessionManager) handleMsg() {
	unpacker := msgpack.PopUnPacker()
	defer msgpack.PushUnPacker(unpacker)
	for _, session := range manager.ssmap {
		session.handleMsg(unpacker)
	}
}

func (manager *SessionManager) Run() {
	manager.handleEvent()
	manager.handleMsg()
}

func (manager *SessionManager) Stop() bool {
	if reslut := manager.server.Stop(); reslut {
		manager.server.SetWatcher(nil)
		manager.ssmap = nil
		return true
	}
	return false
}

//BroadcastMsg 广播消息
func (manager *SessionManager) TraverseAllSession(f func(session ISession)) {
	if f == nil {
		return
	}
	for _, session := range manager.ssmap {
		f(session)
	}
}

//BroadcastMsg 广播消息
func (manager *SessionManager) BroadcastMsg(msg protocolbase.IMsg) {
	packer := msgpack.PopPacker()
	defer msgpack.PushPacker(packer)
	msg.Pack(packer, true)
	data := packer.GetBuffer()
	for _, session := range manager.ssmap {
		session.SendBytes(data)
	}
}

//Kick 踢掉指定连接
func (manager *SessionManager) Kick(id uint64) {
	if session := manager.getSession(id); session != nil {
		session.Close()
	}
}

func (manager *SessionManager) Start() bool {
	if reslut := manager.server.Start(); reslut {
		manager.init()
		manager.ssmap = make(map[uint64]*Session)
		manager.server.SetWatcher(manager)
		return true
	}
	return false
}

func (manager *SessionManager) Count() int {
	return len(manager.ssmap)
}

func (manager *SessionManager) GetSession(id uint64) ISession {
	return manager.getSession(id)
}

func (manager *SessionManager) Name() string {
	return manager.name
}
