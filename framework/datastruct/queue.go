package datastruct

import (
	//"g_server/framework/log"
	"sync"
)

const (
	MaxPool = 1000
)

type queueItem struct {
	value interface{}
	next  *queueItem
}

type Queue struct {
	count   int
	first   *queueItem
	pos     *queueItem
	_first  *queueItem
	_pos    *queueItem
	_pcount int
}

func (queue *Queue) getCount() int {
	return queue.count
}

func (queue *Queue) getItem() *queueItem {
	if queue._first != nil {
		item := queue._first
		queue._first = queue._first.next
		if queue._first == nil {
			queue._pos = nil
		}
		queue._pcount--
		item.value, item.next = nil, nil
		return item
	}
	return &queueItem{}
}

func (queue *Queue) setItem(item *queueItem) {
	if queue._pcount >= MaxPool {
		return
	}
	item.value, item.next = nil, nil
	queue._pcount++
	if queue._first == nil {
		queue._first = item
		queue._pos = queue._first
		return
	}
	queue._pos.next = item
	queue._pos = queue._pos.next
}

/*
func (queue *Queue) PrintPool() {
	item := queue._first
	i := 0
	for {
		if item == nil {
			break
		}
		i++
		glog.LogConsole(glog.LogInfo, "pool index:", i, "value", item.value)
		item = item.next
	}
	glog.LogConsole(glog.LogInfo, "_count _first _pos", queue._pcount, queue._first, queue._pos)
}

func (queue *Queue) Print() {
	item := queue.first
	i := 0
	for {
		if item == nil {
			break
		}
		i++
		glog.LogConsole(glog.LogInfo, "queue index:", i, "value", item.value)
		item = item.next
	}
	glog.LogConsole(glog.LogInfo, "count first pos", queue.count, queue.first, queue.pos)
}
*/
func (queue *Queue) Push(value interface{}) {
	item := queue.getItem()
	item.value = value
	if queue.first == nil {
		queue.first = item
		queue.pos = queue.first
	} else {
		queue.pos.next = item
		queue.pos = queue.pos.next
	}
	queue.count++
}

func (queue *Queue) Pop() interface{} {
	if queue.first == nil {
		return nil
	}
	value, item := queue.first.value, queue.first
	queue.first = queue.first.next
	if queue.first == nil {
		queue.pos = nil
	}
	queue.setItem(item)
	queue.count--
	return value
}

func (queue *Queue) Peek() interface{} {
	if queue.first == nil {
		return nil
	}
	return queue.first.value
}

func (queue *Queue) SwapQueue(copyqueue *Queue) {
	if queue.first == nil {
		return
	}
	copyqueue.first = queue.first
	copyqueue.pos = queue.pos
	copyqueue.count = queue.count

	queue._first = copyqueue._first
	queue._pos = copyqueue._pos
	queue._pcount = copyqueue._pcount

	queue.first = nil
	queue.pos = nil
	queue.count = 0

	copyqueue._first = nil
	copyqueue._pos = nil
	copyqueue._pcount = 0
}

type SyncQueue struct {
	Queue
	mutex sync.RWMutex
}

/*
func (queue *SyncQueue) PrintPool() {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()
	queue.Queue.PrintPool()
}

func (queue *SyncQueue) Print() {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()
	queue.Queue.Print()
}
*/
func (queue *SyncQueue) Push(value interface{}) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.Queue.Push(value)
}

func (queue *SyncQueue) Pop() interface{} {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return queue.Queue.Pop()
}

func (queue *SyncQueue) Peek() interface{} {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()
	return queue.Queue.Peek()
}

func (queue *SyncQueue) SwapQueue(copyqueue *Queue) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.Queue.SwapQueue(copyqueue)
}
