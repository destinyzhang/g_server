package session

import (
	"g_server/framework/com"
	"g_server/framework/gnet"
	"g_server/framework/msgpack"
	"time"
)

type SessionClient struct {
	BaseSession
	SessionMsgProxy
	SessionMsgQueue
	rcontime int32
	state    int32
	name     string
}

func (session *SessionClient) OnSocketMessage(ws gnet.ISocket, msg []byte) {
	session.msgrec.Push(&sessionEvent{event: sessionEventMsg, ws: ws, msgdata: msg})
}

func (session *SessionClient) OnSocketOpen(ws gnet.ISocket) {
	session.msgrec.Push(&sessionEvent{event: sessionEventOpen, ws: ws})
}

func (session *SessionClient) OnSocketClose(ws gnet.ISocket) {
	session.msgrec.Push(&sessionEvent{event: sessionEventClose, ws: ws})
}

func (session *SessionClient) handleEvent() {
	session.copymsg()
	for {
		ievent := session.msghand.Pop()
		if ievent == nil {
			return
		}
		event := ievent.(*sessionEvent)
		switch event.event {
		case sessionEventOpen:
			{
				session.state = 2
				if session.fsessionOpen != nil {
					session.fsessionOpen(session)
				}
			}
		case sessionEventClose:
			{
				session.state = 0
				session.reCon()
				if session.fSessionClose != nil {
					session.fSessionClose(session)
				}
			}
		case sessionEventMsg:
			{
				unpacker := msgpack.PopUnPacker()
				unpacker.Attatch(event.msgdata)
				if r, id := unpacker.UnPackUInt32(); r == 0 {
					if msgProxy := session.FindMsgProxy(id); msgProxy != nil {
						com.SafeCall(func() {
							if msg := msgProxy.msgCreate(); msg != nil {
								msgProxy.msgHandler(session, msg, msg.Unpack(unpacker) == 0)
							}
						})
					}
				}
				msgpack.PushUnPacker(unpacker)
			}
		}
	}

}

func (session *SessionClient) reCon() {
	if session.state == 0 {
		session.state = 1
		go func() {
			for {
				if session.ws.Start() {
					session.ws.SetWatcher(session)
					return
				}
				time.Sleep(time.Duration(session.rcontime) * time.Second)
			}
		}()
	}
}

func (session *SessionClient) Valid() bool {
	return session.state == 2
}

func (session *SessionClient) Run() {
	session.handleEvent()
}

func (session *SessionClient) Start() bool {
	session.init()
	session.reCon()
	return true
}

func (session *SessionClient) Stop() bool {
	session.ws.SetWatcher(nil)
	session.ws.Close()
	return true
}

func (session *SessionClient) Name() string {
	return session.name
}
