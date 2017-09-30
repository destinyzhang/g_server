package datastruct

import (
	"runtime"
	"sync/atomic"
)

type SafeRingQueue struct {
	count   uint32
	mod     uint64
	ring    []interface{}
	pushidx uint64
	popidx  uint64
}

func (queue *SafeRingQueue) init() *SafeRingQueue {
	queue.count = Nextpowof2(queue.count)
	queue.mod = uint64(queue.count) - 1
	queue.ring = make([]interface{}, queue.count)
	return queue
}

func (queue *SafeRingQueue) Cap() uint32 {
	return queue.count
}

func (queue *SafeRingQueue) Push(value interface{}) {
	for {
		pushidx := queue.pushidx
		popidx := queue.popidx
		newpush := pushidx + 1
		if newpush >= popidx+queue.mod {
			runtime.Gosched()
			continue
		}
		if atomic.CompareAndSwapUint64(&queue.pushidx, pushidx, newpush) {
			pos := newpush & queue.mod
			queue.ring[pos] = value
			return
		} else {
			runtime.Gosched()
		}
	}
}

func (queue *SafeRingQueue) Pop() interface{} {
	pushidx := queue.pushidx
	popidx := queue.popidx
	if pushidx <= popidx {
		return nil
	}
	newpop := popidx + 1
	if atomic.CompareAndSwapUint64(&queue.popidx, popidx, newpop) {
		pos := newpop & queue.mod
		value := queue.ring[pos]
		queue.ring[pos] = nil
		return value
	}
	runtime.Gosched()
	return nil
}
