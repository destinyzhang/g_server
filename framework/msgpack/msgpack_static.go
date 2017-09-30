package msgpack

import (
	"bytes"
)

var (
	pool *msgPackUnPackPool
)

func NewPacker() *Packer {
	return &Packer{out_: bytes.NewBuffer(make([]byte, PackbuffSize))}
}

func NewUnPacker() *UnPacker {
	return &UnPacker{in_: bytes.NewBuffer(make([]byte, PackbuffSize))}
}

func InitMsgPackUnPackPool(poolnum int) {
	if pool == nil {
		pool = &msgPackUnPackPool{}
		pool.init(poolnum)
	}
}

func DestoryMsgPackUnPackPool() {
	if pool != nil {
		pool.destory()
	}
}

func PopPacker() *Packer {
	if pool != nil {
		return pool.popPacker()
	}
	return NewPacker()
}

func PushPacker(packer *Packer) {
	if pool != nil {
		pool.pushPacker(packer)
	}
}

func PopUnPacker() *UnPacker {
	if pool != nil {
		return pool.popUnPacker()
	}
	return NewUnPacker()
}

func PushUnPacker(unpacker *UnPacker) {
	if pool != nil {
		pool.pushUnPacker(unpacker)
	}
}
