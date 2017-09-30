package gnet

import (
	"g_server/framework/log"
	"net"
)

//BaseServer 服务器
type BaseServer struct {
	connid   uint64
	host     string
	listen   net.Listener
	state    int
	fnewConn func(conn net.Conn)
	network  string
}

//Stop 关闭
func (server *BaseServer) Stop() bool {
	//这里不会多线程调用三，不管了不加锁
	if server.state == WsServerStateClosed || server.state == WsServerStateCloseing {
		return false
	}
	server.state = WsServerStateCloseing
	err := server.listen.Close()
	server.state = WsServerStateClosed
	if err != nil {
		glog.LogConsole(glog.LogError, "close server err", err)
	}
	return true
}

//Start 开启
func (server *BaseServer) Start() bool {
	listen, err := net.Listen(server.network, server.host)
	if err != nil {
		glog.LogConsole(glog.LogError, "start server fail", err)
		return false
	}
	server.listen = listen
	server.state = WsServerListenning
	server.accept()
	glog.LogConsole(glog.LogInfo, "start server")
	return true
}

func (server *BaseServer) accept() {
	go func() {
		defer server.Stop()
		for {
			conn, err := server.listen.Accept()
			if err != nil {
				glog.LogConsole(glog.LogError, "server accept", err)
				return
			}
			if server.state != WsServerListenning {
				return
			}
			if server.fnewConn != nil {
				server.fnewConn(conn)
			}
		}
	}()
}

func (server *BaseServer) genConnid() uint64 {
	server.connid++
	return server.connid
}
