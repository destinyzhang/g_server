package glog

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	_logLevel = 0
	_logMutex = new(sync.Mutex)
)

const (
	LogForce    = 0
	LogError    = 1 << 0
	LogWarning  = 1 << 1
	LogInfo     = 1 << 2
	LenStackBuf = 1024
)

//LogConsole 打印
func LogConsole(lv int, log ...interface{}) {
	if lv == LogForce || _logLevel&lv > 0 {
		_logMutex.Lock()
		defer _logMutex.Unlock()
		if lv == LogError {
			buf := make([]byte, LenStackBuf)
			log = append(log, string(buf[:runtime.Stack(buf, false)]))
		}
		fmt.Println(log...)
	}
}

//SetLog 设置
func SetLog(level int) {
	_logLevel = level
	fmt.Println("log state:", _logLevel)
}
