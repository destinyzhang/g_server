package datastruct

func Ispowof2(num uint32) bool {
	return num&(num-1) == 0
}

func Nextpowof2(num uint32) uint32 {
	if Ispowof2(num) {
		return num
	}
	for i := 0; i < 4; i++ {
		num |= num >> (1 << uint32(i))
	}
	return num + 1
}

func NewSafeRingQueue(count uint32) *SafeRingQueue {
	return (&SafeRingQueue{count: count}).init()
}

func NewQueue() *Queue {
	return &Queue{}
}

func NewSyncQueue() *SyncQueue {
	return &SyncQueue{}
}
