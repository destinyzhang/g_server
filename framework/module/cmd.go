package module

import (
	"bufio"
	"flag"
	"g_server/framework/com"
	"g_server/framework/log"
	"net"
	"os"
	"strings"
)

type cmdItem struct {
	cmd   string
	parms []string
	addr  net.Addr
}

type cmdHandler struct {
	handler func([]string) (bool, string)
	des     string
	remote  bool
}

type CmdDispatcher struct {
	iplist   []string
	listen   net.PacketConn
	cmdchan  chan *cmdItem
	handlers map[string]*cmdHandler
	run      bool
	name     string
}

func (disp *CmdDispatcher) RegCmd(name string, f func([]string) (bool, string), des string, remote bool) bool {
	if _, ok := disp.handlers[name]; ok == true {
		return false
	}
	disp.handlers[name] = &cmdHandler{handler: f, des: des, remote: remote}
	return true
}

func (disp *CmdDispatcher) checkip(ip string) bool {
	glog.LogConsole(glog.LogInfo, "checkip :", ip)
	for _, v := range disp.iplist {
		glog.LogConsole(glog.LogInfo, "checkip v :", v)
		if strings.Contains(ip, v) {
			return true
		}
	}
	return false
}

func (disp *CmdDispatcher) replayeRemoteCmd(txt string, addr net.Addr) {
	com.SafeCall(func() {
		if disp.listen != nil {
			disp.listen.WriteTo([]byte(txt), addr)
		}
	})
}

func (disp *CmdDispatcher) acceptRemote() {
	go func() {
		defer func() {
			disp.listen.Close()
			disp.listen = nil
		}()
		var (
			cmd   string
			parms string
		)
		flagset := flag.NewFlagSet("CmdRemoteDispatcher", flag.ContinueOnError)
		flagset.StringVar(&cmd, "cmd", "", "cmd name")
		flagset.StringVar(&parms, "parms", "", "parms split by ,")
		buffer := make([]byte, 1024)
		for {
			n, addr, err := disp.listen.ReadFrom(buffer[0:])
			if err != nil {
				glog.LogConsole(glog.LogError, "CmdRemoteDispatcher read err:", err)
				return
			}
			com.SafeCall(func() {
				if !disp.checkip(addr.String()) {
					disp.replayeRemoteCmd("fuck!!", addr)
					return
				}

				line := string(buffer[:n])
				if err := flagset.Parse(strings.Split(line, " ")); err != nil {
					disp.replayeRemoteCmd(err.Error(), addr)
					return
				}

				if cmd == "" {
					disp.replayeRemoteCmd("cmd empty!!", addr)
					return
				}
				disp.cmdchan <- &cmdItem{cmd: cmd, parms: strings.Split(string(parms), ","), addr: addr}
			})
		}

	}()
}

func (disp *CmdDispatcher) recvCmd() {
	go func() {
		var (
			cmd   string
			parms string
		)
		flagset := flag.NewFlagSet("CmdDispatcher", flag.ContinueOnError)
		flagset.StringVar(&cmd, "cmd", "", "cmd name")
		flagset.StringVar(&parms, "parms", "", "parms split by ,")
		reader := bufio.NewReader(os.Stdin)
		for {
			if !disp.run {
				return
			}
			com.SafeCall(func() {
				data, _, _ := reader.ReadLine()
				line := string(data)
				if err := flagset.Parse(strings.Split(line, " ")); err != nil {
					glog.LogConsole(glog.LogError, "cmdDispatcher parse err:", err)
					return
				}
				if cmd == "" {
					return
				}
				disp.cmdchan <- &cmdItem{cmd: cmd, parms: strings.Split(string(parms), ",")}
			})
		}
	}()
}
func (disp *CmdDispatcher) SetRemoteIp(ips []string, clear bool) {
	if clear {
		disp.iplist = make([]string, 0, 10)
	}
	disp.iplist = append(disp.iplist, ips...)
	glog.LogConsole(glog.LogInfo, " disp.iplist ", disp.iplist)
}

func (disp *CmdDispatcher) Run() {
	if disp.run {
		select {
		case cmd, ok := <-disp.cmdchan:
			{
				if ok {
					com.SafeCall(func() {
						if handler, ok := disp.handlers[cmd.cmd]; ok {
							result, info := handler.handler(cmd.parms)
							if cmd.addr != nil {
								disp.replayeRemoteCmd("["+cmd.cmd+"]["+com.ConverToStr(result)+"]["+info+"]", cmd.addr)
							}
							glog.LogConsole(glog.LogForce, "do cmd :", cmd.cmd, result, info)
						}
					})
				}
			}
		default:
		}
	}
}

func (disp *CmdDispatcher) Start() bool {
	if !disp.run {
		disp.cmdchan = make(chan *cmdItem)
		disp.recvCmd()
		disp.run = true
	}
	return true
}

func (disp *CmdDispatcher) StartRemote(host string) bool {
	if disp.listen != nil {
		return true
	}
	con, err := net.ListenPacket("udp", host)
	if err != nil {
		glog.LogConsole(glog.LogError, "cmd remote start err", err)
		return false
	}
	disp.listen = con
	disp.acceptRemote()
	return true
}

func (disp *CmdDispatcher) CloseRemote() {
	com.SafeCall(func() {
		if disp.listen == nil {
			return
		}
		disp.listen.Close()
	})
}

func (disp *CmdDispatcher) Stop() bool {
	if disp.run {
		disp.run = false
		close(disp.cmdchan)
	}

	return true
}

func (disp *CmdDispatcher) Name() string {
	return disp.name
}
