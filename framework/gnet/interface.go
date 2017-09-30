package gnet

type ISocket interface {
	TypeName() string
	LocalAddr() string
	RemoteAddr() string
	ID() uint64
	State() int
	Close() bool
	Start() bool
	SendBit([]byte)
	SetWatcher(ISocketWatcher)
	GetWatcher() ISocketWatcher
}

type ISocketWatcher interface {
	OnSocketOpen(ISocket)
	OnSocketClose(ISocket)
	OnSocketMessage(ISocket, []byte)
}

type IServer interface {
	TypeName() string
	Stop() bool
	Start() bool
	SetMaxMsgSize(uint32)
	GetMaxMsgSize() uint32
	SetWatcher(IServerWatcher)
	GetWatcher() IServerWatcher
}

type IServerWatcher interface {
	OnSocketAccept(ISocket)
}
