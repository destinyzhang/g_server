package msgpack

import (
	"g_server/framework/com"
)

type msgPackUnPackPool struct {
	packChan   chan *Packer
	unpackChan chan *UnPacker
	run        bool
}

func (pool *msgPackUnPackPool) destory() {
	if !pool.run {
		return
	}
	pool.run = false
	close(pool.packChan)
	close(pool.unpackChan)
}

func (pool *msgPackUnPackPool) init(pnum int) {
	if pool.run {
		return
	}
	pool.run = true
	pool.packChan = make(chan *Packer, pnum)
	pool.unpackChan = make(chan *UnPacker, pnum)
}

func (pool *msgPackUnPackPool) pushPacker(packer *Packer) {
	if pool.run {
		com.SafeCall(func() {
			select {
			case pool.packChan <- packer:
			default:
				return
			}
		})

	}
}

func (pool *msgPackUnPackPool) pushUnPacker(unpacker *UnPacker) {
	if pool.run {
		com.SafeCall(func() {
			select {
			case pool.unpackChan <- unpacker:
			default:
				return
			}
		})
	}
}

func (pool *msgPackUnPackPool) popPacker() *Packer {
	if pool.run {
		select {
		case packer, ok := <-pool.packChan:
			if ok {
				return packer
			}
		default:
		}
	}
	return NewPacker()
}

func (pool *msgPackUnPackPool) popUnPacker() *UnPacker {
	if pool.run {
		select {
		case unpacker, ok := <-pool.unpackChan:
			if ok {
				return unpacker
			}
		default:
		}
	}
	return NewUnPacker()
}
