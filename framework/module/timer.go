package module

import (
	"container/list"
	"g_server/framework/com"
	"g_server/framework/log"
	"sync"
	"time"
)

const (
	_timerSlotTick = uint64(10)                           //100毫秒检测一次
	_timerSlot     = uint64(3600 * 1000 / _timerSlotTick) //1圈槽的个数刚好1个小时
)

//Timer 计时器
type ItimerCallBack interface {
	Callback() bool
}

//Timer 计时器
type Timer struct {
	needround uint64
	slot      uint64
	callback  ItimerCallBack
	aftertime uint64
	valid     bool
}

func (timer *Timer) doCallBack(disp *TimerDispatcher) {
	if timer.callback != nil {
		com.SafeCall(func() {
			if timer.callback.Callback() {
				disp.addTimer(timer)
			}
		})
	}
}

//TimerDispatcher 调度结构
type TimerDispatcher struct {
	slotlist  [_timerSlot]*list.List //超时槽
	tickcount uint64                 //tick次数
	runtime   int64
	name      string
}

func (disp *TimerDispatcher) add2Slot(timer *Timer) {
	clist := disp.slotlist[timer.slot] //取槽里的数据
	if clist == nil {
		clist = new(list.List)
		disp.slotlist[timer.slot] = clist
	}
	clist.PushBack(timer)
}

func (disp *TimerDispatcher) addTimer(timer *Timer) (*Timer, bool) {
	slot := timer.aftertime / _timerSlotTick //需要多少槽
	if slot > 0 {
		slot -= 1
	}
	timer.needround = slot / _timerSlot //在那一圈
	timer.slot = (slot + disp.tickcount%_timerSlot) % _timerSlot
	disp.add2Slot(timer)
	glog.LogConsole(glog.LogInfo, "add timer", timer.needround, timer.aftertime)
	return timer, true
}

//AddTimer 加入一个timer
func (disp *TimerDispatcher) AddTimer(callback ItimerCallBack, aftertime uint64) (*Timer, bool) {
	return disp.addTimer(&Timer{callback: callback, aftertime: aftertime, valid: true})
}

func (disp *TimerDispatcher) pickRunTimer() (callTimer []*Timer) {
	curslot := disp.tickcount % _timerSlot //当前那个槽
	disp.tickcount++
	clist := disp.slotlist[curslot] //取槽里的数据
	if clist != nil {
		for it := clist.Front(); it != nil; {
			cit := it
			it = it.Next() //指向下一个，因为可能要删除
			timer := cit.Value.(*Timer)
			if !timer.valid { //非法timer删除了
				clist.Remove(cit)
				continue
			}
			if timer.needround > 0 { //当前圈不执行
				timer.needround--
				continue
			}
			//当前圈执行
			clist.Remove(cit)
			callTimer = append(callTimer, timer)
		}
	}
	return
}

//Cancel 取消一个timer
func (disp *TimerDispatcher) Cancel(timer *Timer) {
	timer.valid = false //设为不可用就行了，没必要再去删除,没必要加锁run里面只是读他的值
}

func (disp *TimerDispatcher) Run() {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-disp.runtime >= int64(_timerSlotTick) {
		disp.runtime = now
		disp.run()
	}
}

func (disp *TimerDispatcher) run() {
	callTimer := disp.pickRunTimer()
	for _, timer := range callTimer {
		timer.doCallBack(disp)
	}
}

//Start 开启调度
func (disp *TimerDispatcher) Start() bool {
	disp.runtime = 0
	return true
}

//Stop 关闭
func (disp *TimerDispatcher) Stop() bool {
	return true
}

func (disp *TimerDispatcher) Name() string {
	return disp.name
}

//SyncTimerDispatcher 调度结构
type SyncTimerDispatcher struct {
	TimerDispatcher
	sync.Mutex
	runing bool
}

//IsRun 运行状态
func (disp *SyncTimerDispatcher) IsRun() bool {
	return disp.runing
}

func (disp *SyncTimerDispatcher) add2Slot(timer *Timer) {
	disp.Lock()
	defer disp.Unlock()
	disp.TimerDispatcher.add2Slot(timer)
}

func (disp *SyncTimerDispatcher) pickRunTimer() (callTimer []*Timer) {
	disp.Lock()
	defer disp.Unlock()
	callTimer = disp.TimerDispatcher.pickRunTimer()
	return
}

//Start 开启调度
func (disp *SyncTimerDispatcher) Start() bool {
	if disp.TimerDispatcher.Start() {
		go func() {
			glog.LogConsole(glog.LogInfo, "SyncTimerDispatcher start")
			tick := time.NewTicker(time.Duration(_timerSlotTick) * time.Millisecond)
			defer tick.Stop()
			for disp.runing {
				<-tick.C
				disp.run()
			}
			glog.LogConsole(glog.LogInfo, "SyncTimerDispatcher close")
		}()
	}
	return disp.runing
}
