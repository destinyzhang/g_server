package com

import (
	"g_server/framework/log"
)

func SafeGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				glog.LogConsole(glog.LogError, r)
			}
		}()
		f()
	}()
}

func SafeCall(f func()) {
	defer func() {
		if r := recover(); r != nil {
			glog.LogConsole(glog.LogError, r)
		}
	}()
	f()
}
