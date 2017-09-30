package gnet

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"g_server/framework/log"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"
)

//WebSocket 基础类
type WebSocket struct {
	conn         net.Conn
	rw           *bufio.ReadWriter
	needmask     bool
	state        int
	connid       uint64
	sendchan     chan *webSocketMsg
	sendbuff     []byte
	readbuff     []byte
	readbuffPos  uint32
	codeMask     [8]byte
	wsmaxmsgsize uint32
	path         string
	watcher      ISocketWatcher
}

func (ws *WebSocket) TypeName() string {
	return "websocket"
}

func (ws *WebSocket) Path() string {
	return ws.path
}

func (ws *WebSocket) LocalAddr() string {
	return ws.conn.LocalAddr().String()
}

func (ws *WebSocket) RemoteAddr() string {
	return ws.conn.RemoteAddr().String()
}

//SetMaxMsgSize 设置接受最大包大小
func (ws *WebSocket) SetMaxMsgSize(size uint32) {
	ws.wsmaxmsgsize = size
}

//GetMaxMsgSize 返回接受最大包大小
func (ws *WebSocket) GetMaxMsgSize() uint32 {
	return ws.wsmaxmsgsize
}

//SetWatcher
func (ws *WebSocket) SetWatcher(watcher ISocketWatcher) {
	ws.watcher = watcher
}

//GetWatcher
func (ws *WebSocket) GetWatcher() ISocketWatcher {
	return ws.watcher
}

//ID 返回ID
func (ws *WebSocket) ID() uint64 {
	return ws.connid
}

//State 返回状态
func (ws *WebSocket) State() int {
	return ws.state
}

//Close 关闭连接
func (ws *WebSocket) Close() bool {
	glog.LogConsole(glog.LogInfo, "state", ws.state)
	if ws.state == WsStateClosed || ws.state == WsStateCloseing {
		return true
	}
	ws.pushMsgChan(_wsOpcodeClose, []byte("close"))
	ws.state = WsStateCloseing
	close(ws.sendchan)
	return true
}

//Start 开始函数
func (ws *WebSocket) Start() bool {
	go func() {
		if !ws.accepthandshake() {
			ws.state = WsStateClosed
			ws.conn.Close()
			glog.LogConsole(glog.LogError, "handshake fail")
			return
		}
		ws.beginSend()
		ws.beginRecv()
	}()
	return true
}

//SendText 发送字符串
func (ws *WebSocket) SendText(data string) {
	ws.pushMsgChan(_wsOpcodeTxt, []byte(data))
}

//SendBit 发送二进制
func (ws *WebSocket) SendBit(data []byte) {
	//拷贝一份避免外面修改slice
	_data := make([]byte, len(data))
	copy(_data, data)
	ws.pushMsgChan(_wsOpcodeBit, _data)
}

//Ping 发送Ping
func (ws *WebSocket) Ping() {
	ws.pushMsgChan(_wsOpcodePing, []byte("ping"))
}

//Pong 发送Pong
func (ws *WebSocket) Pong() {
	ws.pushMsgChan(_wsOpcodePong, []byte("pong"))
}

func (ws *WebSocket) remakeReadBuff(size uint32) {
	readbuff := make([]byte, size)
	copy(readbuff, ws.readbuff)
	ws.readbuff = readbuff
}

func (ws *WebSocket) write(data []byte) error {
	//ws.conn.SetWriteDeadline((time.Now().Add(time.Second * 5)))
	_, err := ws.conn.Write(data)
	if err != nil {
		glog.LogConsole(glog.LogError, "write err:", err)
		return err
	}
	return nil
}

func (ws *WebSocket) readBuffer(buff []byte) (int, error) {
	return io.ReadFull(ws.rw, buff)
}

func (ws *WebSocket) read() ([]byte, error) {
	//ws.conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
	n, err := ws.conn.Read(ws.readbuff)
	if err != nil {
		glog.LogConsole(glog.LogError, "read err:", err)
		return nil, err
	}
	return ws.readbuff[:n], err
}

//SendMsg 发送消息
func (ws *WebSocket) pushMsgChan(opcode byte, data []byte) {
	if ws.state == WsStateClosed || ws.state == WsStateCloseing {
		return
	}
	ws.sendchan <- &webSocketMsg{opcode: opcode, buff: data}
}

//sendMsg 发送消息
func (ws *WebSocket) sendMsg(opcode byte, data []byte) error {
	if datalen := len(data); datalen > _buffCap {
		frag := datalen / _buffCap
		left := datalen % _buffCap
		for i := 0; i < frag; i++ {
			if i != 0 {
				opcode = _wsOpcodeCon
			}
			if err := ws.sendFrame(left == 0 && i == frag-1, opcode, data[i*_buffCap:(i+1)*_buffCap]); err != nil {
				return err
			}
		}
		if left > 0 {
			return ws.sendFrame(true, _wsOpcodeCon, data[frag*_buffCap:])
		}
		return nil
	}
	return ws.sendFrame(true, opcode, data)
}

//Genbase64key 生成长度klen的base64字符串
func (ws *WebSocket) genbase64key(klen int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < klen; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return base64.StdEncoding.EncodeToString(result)
}

func (ws *WebSocket) createMaskingKey(klen int) []byte {
	rand.Seed(time.Now().UnixNano())
	mask := make([]byte, klen, klen)
	for i := 0; i < klen; i++ {
		mask[i] = byte(rand.Intn(124) + 1)
	}
	return mask
}

func (ws *WebSocket) sendFrame(end bool, opcode byte, data []byte) (err error) {
	length := len(data)
	buf := ws.sendbuff[0:0]
	var finBit, maskBit byte
	if end {
		finBit = 0x80
	} else {
		finBit = 0
	}
	buf = append(buf, finBit|opcode)
	//掩码标记
	if ws.needmask {
		maskBit = 0x80
	} else {
		maskBit = 0
	}
	//长度写入
	if length < 126 {
		buf = append(buf, byte(length)|maskBit)
	} else if length < 0xffff {
		buf = append(buf, 126|maskBit, 0, 0)
		binary.BigEndian.PutUint16(buf[len(buf)-2:], uint16(length))
	} else {
		buf = append(buf, 127|maskBit, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(buf[len(buf)-8:], uint64(length))
	}
	//写入掩码
	if ws.needmask {
		codeMask := ws.createMaskingKey(4)
		buf = append(buf, codeMask...)
		ws.mask(codeMask, data)
	}
	//写入数据
	buf = append(buf, data...)
	_, err = ws.rw.Write(buf)
	if err != nil {
		return
	}
	err = ws.rw.Flush()
	return
}

func (ws *WebSocket) mask(mask []byte, data []byte) {
	for i := range data {
		data[i] ^= mask[i%4]
	}
}

func (ws *WebSocket) recvFrame() (opcode byte, finl bool, err error) {
	buff := ws.codeMask[0:]
	//读取前两位F RRR  opcode
	_, err = ws.readBuffer(buff[:2])
	if err != nil {
		return
	}
	header, payload := buff[0], buff[1]
	finl = header&0x80 != 0 //是否是结束帧
	//判断扩展是否为0，不为0就是错误的
	if header&0x70 != 0 {
		err = ErrRSV123
		return
	}
	opcode = header & 0xf //opcode的值
	switch opcode {
	case _wsOpcodeCon:
	case _wsOpcodeTxt:
	case _wsOpcodeBit:
	case _wsOpcodeClose:
	case _wsOpcodePing:
	case _wsOpcodePong:
	default:
		{
			err = ErrInvalidOpcode
			return
		}
	}
	maskFrame := payload&0x80 != 0       //是否掩码
	payloadlen := uint32(payload & 0x7f) //payload长度
	switch {
	case payloadlen == 126: //后面2个字节16位无符号，网络字节序
		{
			_, err = ws.readBuffer(buff[:2])
			if err != nil {
				return
			}
			payloadlen = uint32(binary.BigEndian.Uint16(buff[:2]))
		}
	case payloadlen == 127: //后面8个字节64位无符号，网络字节序
		{
			_, err = ws.readBuffer(buff[:8])
			if err != nil {
				return
			}
			payloadlen = binary.BigEndian.Uint32(buff[:8])
		}
	}

	buffneed := ws.readbuffPos + payloadlen
	if buffneed > ws.wsmaxmsgsize {
		err = ErrMsgSizeInvalid
		return
	}

	codeMask := ws.codeMask[0:4]
	//取掩码
	if maskFrame {
		_, err = ws.readBuffer(codeMask)
		if err != nil {
			return
		}
	}
	//长度很长需要重新分配
	if buffneed > uint32(len(ws.readbuff)) {
		ws.remakeReadBuff(buffneed * 2)
	}
	//取数据
	appbuff := ws.readbuff[ws.readbuffPos : ws.readbuffPos+payloadlen]
	_, err = ws.readBuffer(appbuff)
	if err != nil {
		return
	}
	//去掉掩码
	if maskFrame {
		ws.mask(codeMask, appbuff)
	}
	ws.readbuffPos = buffneed
	return
}

func (ws *WebSocket) parsehandshake(porlStr string) (kvmap map[string]string) {
	glog.LogConsole(glog.LogInfo, "parsehandshake:", porlStr)
	kvmap = make(map[string]string)
	for i, v := range strings.Split(porlStr, "\r\n") {
		idx := strings.Index(v, ":")
		if idx < 0 {
			if i == 0 {
				kvmap[_wsHkHead] = v
			}
			continue
		}
		kvmap[v[:idx]] = strings.TrimSpace(v[idx+1:])
	}
	glog.LogConsole(glog.LogInfo, "kvmap:", kvmap)
	return
}

func (ws *WebSocket) acceptKey(hkKey string) (base64str string) {
	ha1 := sha1.New()
	ha1.Write([]byte(hkKey))
	ha1.Write(_wsMagicKey)
	base64str = base64.StdEncoding.EncodeToString(ha1.Sum(nil))
	return
}

func (ws *WebSocket) accepthandshake() bool {
	ws.readbuff = make([]byte, _buffCap)
	buff, err := ws.read()
	if err != nil {
		return false
	}
	porlStr := string(buff)
	kvmap := ws.parsehandshake(porlStr)
	if kvmap[_wsHkUpgrade] != "websocket" {
		glog.LogConsole(glog.LogWarning, "_ws_hkUpgrade:", kvmap[_wsHkUpgrade])
		return false
	}

	if kvmap[_wsHkVersion] != "13" {
		glog.LogConsole(glog.LogWarning, "_ws_hkVersion:", kvmap[_wsHkVersion])
		return false
	}
	hkKey, ok := kvmap[_wsHkKey]
	if !ok {
		glog.LogConsole(glog.LogWarning, "_ws_hkKey no exist")
		return false
	}

	if head, ok := kvmap[_wsHkHead]; ok {
		heads := strings.Split(head, " ")
		if len(heads) == 3 {
			ws.path = heads[1]
		}
	}

	base64str := ws.acceptKey(hkKey)
	buf := bytes.NewBufferString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: ")
	buf.WriteString(base64str)
	buf.Write(_wsCrlf)
	buf.Write(_wsCrlf)

	if err := ws.write(buf.Bytes()); err == nil {
		return true
	}
	return false
}

func (ws *WebSocket) beginRecv() {
	go func() {
		defer func() {
			glog.LogConsole(glog.LogInfo, "beginRecv end")
			ws.Close()
			ws.readbuff = nil
			ws.readbuffPos = 0
		}()
		ws.state = WsStateConnected
		if ws.watcher != nil {
			ws.watcher.OnSocketOpen(ws)
		}
		for {
			var opcode byte = _wsOpcodeCon
			ws.readbuffPos = 0
			for {
				opc, fi, err := ws.recvFrame()
				if err != nil {
					glog.LogConsole(glog.LogError, "recvFrame:", err)
					return
				}
				if opc != _wsOpcodeCon {
					opcode = opc
				}
				if fi {
					break
				}
			}
			switch opcode {
			case _wsOpcodePong:
			case _wsOpcodeClose:
				ws.Close()
				return
			case _wsOpcodePing:
				ws.Pong()
			case _wsOpcodeTxt:
				fallthrough
			case _wsOpcodeBit:
				{
					if ws.watcher != nil {
						buff := make([]byte, ws.readbuffPos)
						copy(buff, ws.readbuff[:ws.readbuffPos])
						ws.watcher.OnSocketMessage(ws, buff)
					} else {
						glog.LogConsole(glog.LogInfo, "recv: len=", ws.readbuffPos, " opcode=", opcode)
					}
				}
			}
		}
	}()
}

func (ws *WebSocket) beginSend() {
	ws.sendbuff = make([]byte, _buffCap+_buffCapHead)
	ws.sendchan = make(chan *webSocketMsg, 100)
	go func() {
		defer func() {
			glog.LogConsole(glog.LogInfo, "beginSend end")
			ws.close()
			ws.sendchan = nil
			ws.sendbuff = nil
		}()
		for {
			msg, ok := <-ws.sendchan
			if !ok {
				return
			}
			ws.sendMsg(msg.opcode, msg.buff)
		}
	}()
}

//Close 关闭连接
func (ws *WebSocket) close() {
	err := ws.conn.Close()
	ws.state = WsStateClosed
	if ws.watcher != nil {
		ws.watcher.OnSocketClose(ws)
	}
	glog.LogConsole(glog.LogInfo, "close WebSocket:", err)
}
