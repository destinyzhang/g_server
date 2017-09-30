package module

import (
	"g_server/framework/com"
)

type IAsynItem interface {
	AsynDo()
	Callback()
}

type AsynHelper struct {
	queue chan IAsynItem
	name  string
}

func (help *AsynHelper) AsynDo(item IAsynItem) {
	com.SafeGo(func() {
		item.AsynDo()
		help.queue <- item
	})
}

func (help *AsynHelper) Start() bool {
	return true
}

func (help *AsynHelper) Stop() bool {
	com.SafeCall(func() {
		close(help.queue)
		help.queue = nil
	})
	return true
}

func (help *AsynHelper) Run() {
	select {
	case item, ok := <-help.queue:
		{
			if ok {
				com.SafeCall(func() {
					item.Callback()
				})
			}
		}
	default:
		return
	}
}

func (help *AsynHelper) Name() string {
	return help.name
}
