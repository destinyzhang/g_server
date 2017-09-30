package session

import (
	"g_server/framework/datastruct"
	"g_server/framework/gnet"
	"g_server/framework/msgpack"
	"g_server/framework/protocolbase"
)

type ISession interface {
	ID() uint64
	SendMsg(protocolbase.IMsg)
	SendBytes([]byte)
	IpStr() string
	SetTag(interface{})
	GetTag() interface{}
}

type sessionEvent struct {
	event   int
	msgdata []byte
	ws      gnet.ISocket
}

type msgProxy struct {
	msgHandler func(ISession, protocolbase.IMsg, bool)
	msgCreate  func() protocolbase.IMsg
}

type SessionMsgProxy struct {
	msghanders    map[uint32]*msgProxy
	fsessionOpen  func(ISession)
	fSessionClose func(ISession)
}

func (proxy *SessionMsgProxy) FindMsgProxy(msgid uint32) *msgProxy {
	if msgProxy, ok := proxy.msghanders[msgid]; ok {
		return msgProxy
	}
	return nil
}

func (proxy *SessionMsgProxy) RegIMsgHandler(msgid uint32, mh func(ISession, protocolbase.IMsg, bool), mc func() protocolbase.IMsg) {
	proxy.msghanders[msgid] = &msgProxy{msgHandler: mh, msgCreate: mc}
}

func (proxy *SessionMsgProxy) RegSessionOpen(f func(ISession)) {
	proxy.fsessionOpen = f
}

func (proxy *SessionMsgProxy) RegSessionClose(f func(ISession)) {
	proxy.fSessionClose = f
}

type SessionMsgQueue struct {
	msgrec  *datastruct.SyncQueue
	msghand *datastruct.Queue
}

func (queue *SessionMsgQueue) init() {
	queue.msgrec = datastruct.NewSyncQueue()
	queue.msghand = datastruct.NewQueue()
}

func (queue *SessionMsgQueue) copymsg() {
	queue.msgrec.SwapQueue(queue.msghand)
}

type BaseSession struct {
	ws  gnet.ISocket
	tag interface{}
}

func (session *BaseSession) SetTag(tag interface{}) {
	session.tag = tag
}

func (session *BaseSession) GetTag() interface{} {
	return session.tag
}

func (session *BaseSession) ID() uint64 {
	return session.ws.ID()
}

func (session *BaseSession) SendMsg(msg protocolbase.IMsg) {
	packer := msgpack.PopPacker()
	defer msgpack.PushPacker(packer)
	msg.Pack(packer, true)
	data := packer.GetBuffer()
	session.SendBytes(data)
}

func (session *BaseSession) SendBytes(data []byte) {
	session.ws.SendBit(data)
}

func (session *BaseSession) Close() {
	session.ws.Close()
}

func (session *BaseSession) IpStr() string {
	return session.ws.RemoteAddr()
}
