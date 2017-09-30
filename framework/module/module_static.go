package module

func NewAsynHelper(name string) *AsynHelper {
	return &AsynHelper{name: name, queue: make(chan IAsynItem, 10)}
}

func NewTimerDispatcher(name string) *TimerDispatcher {
	return &TimerDispatcher{name: name}
}

func NewSyncTimerDispatcher(name string) *SyncTimerDispatcher {
	return &SyncTimerDispatcher{TimerDispatcher: TimerDispatcher{name: name}}
}

func NewCmdDispatcher(name string) *CmdDispatcher {
	return &CmdDispatcher{handlers: make(map[string]*cmdHandler), name: name}
}
