package protocolbase

type IMsg interface {
	GetProId() uint32
	Pack(IPacker, bool)
	Unpack(IUnpacker) int
}

type IPacker interface {
	ClearBuffer()
	PackFloat(float32)
	PackDouble(float64)
	PackBool(bool)
	PackString(string)
	PackInt32(int32)
	PackInt64(int64)
	PackUInt32(uint32)
	PackUInt64(uint64)
	PackBytes([]byte)
}

type IUnpacker interface {
	Attatch([]byte)
	UnPackUInt64() (int, uint64)
	UnPackUInt32() (int, uint32)
	UnPackInt32() (int, int32)
	UnPackInt64() (int, int64)
	UnPackString() (int, string)
	UnPackFloat() (int, float32)
	UnPackDouble() (int, float64)
	UnPackBool() (int, bool)
	UnPackBytes() (int, []byte)
}
