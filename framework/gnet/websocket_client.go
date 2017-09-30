package gnet

import (
	"bufio"
	"bytes"
	"g_server/framework/log"
	"net"
	"net/url"
)

//WebSocketClient 客户端
type WebSocketClient struct {
	WebSocket
	base64key string
	hosturl   string
	network   string
}

func (ws *WebSocketClient) clienthandshake() (err error) {
	u, err := url.Parse(ws.hosturl)
	if err != nil {
		glog.LogConsole(glog.LogError, "url fail", err)
		return
	}
	ws.conn, err = net.Dial(ws.network, u.Host)
	if err != nil {
		glog.LogConsole(glog.LogError, "dial fail", err)
		return
	}

	ws.needmask = true
	ws.base64key = ws.genbase64key(14)
	ws.rw = bufio.NewReadWriter(bufio.NewReader(ws.conn), bufio.NewWriter(ws.conn))

	ws.state = WsStateConnecting
	buf := bytes.NewBufferString("GET ")
	buf.WriteString(u.Path)
	buf.WriteString(" HTTP/1.1\r\n")
	buf.WriteString("Upgrade: websocket\r\nConnection: Upgrade\r\n")
	buf.WriteString("Host: ")
	buf.WriteString(u.Host)
	buf.Write(_wsCrlf)
	buf.WriteString("Sec-WebSocket-Version: 13\r\nSec-WebSocket-Key: ")
	buf.WriteString(ws.base64key)
	buf.Write(_wsCrlf)
	buf.Write(_wsCrlf)

	err = ws.write(buf.Bytes())

	if err != nil {
		glog.LogConsole(glog.LogError, "clienthandshake write fail", err)
		return
	}
	ws.readbuff = make([]byte, _buffCap)
	buff, err := ws.read()
	if err != nil {
		return
	}
	porlStr := string(buff)
	kvmap := ws.parsehandshake(porlStr)
	if kvmap[_wsHkAccept] == "" {
		err = ErrHandshakeEmpty
		return
	}
	if kvmap[_wsHkAccept] != ws.acceptKey(ws.base64key) {
		err = ErrHandshake
		return
	}
	return
}

//Start 客户端连接
func (ws *WebSocketClient) Start() bool {
	err := ws.clienthandshake()
	if err != nil {
		if ws.conn != nil {
			ws.state = WsStateClosed
			ws.conn.Close()
		}
		glog.LogConsole(glog.LogError, "clienthandshake fail", err)
		return false
	}
	ws.beginSend()
	ws.beginRecv()
	glog.LogConsole(glog.LogInfo, "clienthandshake success")
	return true
}
