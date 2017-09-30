package msgpack

import (
	"bytes"
	"encoding/binary"
	"g_server/framework/log"
	"math"
)

const (
	//! Integers
	MP_INT8            = uint8(0xd0)
	MP_INT16           = uint8(0xd1)
	MP_INT32           = uint8(0xd2)
	MP_INT64           = uint8(0xd3)
	MP_UINT8           = uint8(0xcc)
	MP_UINT16          = uint8(0xcd)
	MP_UINT32          = uint8(0xce)
	MP_UINT64          = uint8(0xcf)
	MP_FIXNUM          = uint8(0x00) //!< Last 7 bits is value
	MP_NEGATIVE_FIXNUM = uint8(0xe0) //!< Last 5 bits is value

	//! nil
	MP_NULL = uint8(0xc0)

	//! pre-defined struct
	MP_STRUCT = uint8(0xc1)

	//! boolean
	MP_FALSE = uint8(0xc2)
	MP_TRUE  = uint8(0xc3)

	//! Floating point
	MP_FLOAT  = uint8(0xca)
	MP_DOUBLE = uint8(0xcb)

	/*****************************************************
	 * Variable length types
	 *****************************************************/

	//! Raw bytes
	MP_RAW16  = uint8(0xda)
	MP_RAW32  = uint8(0xdb)
	MP_FIXRAW = uint8(0xa0) //!< Last 5 bits is size

	/*****************************************************
	 * Container types
	 *****************************************************/

	//! Arrays
	MP_ARRAY16  = uint8(0xdc)
	MP_ARRAY32  = uint8(0xdd)
	MP_FIXARRAY = uint8(0x90) //<! Lst 4 bits is size

	//! Maps
	MP_MAP16  = uint8(0xde)
	MP_MAP32  = uint8(0xdf)
	MP_FIXMAP = uint8(0x80) //<! Last 4 bits is size

	//! Some helper bitmasks
	MAX_4BIT  = uint32(0xf)
	MAX_5BIT  = uint32(0x1f)
	MAX_7BIT  = uint32(0x7f)
	MAX_8BIT  = uint32(0xff)
	MAX_15BIT = uint32(0x7fff)
	MAX_16BIT = uint32(0xffff)
	MAX_31BIT = uint32(0x7fffffff)
	MAX_32BIT = uint32(0xffffffff)

	PackbuffSize = 512
)

var (
	USE_BIGENDIAN = false
)

type Packer struct {
	out_ *bytes.Buffer
}

func (packer *Packer) write(i interface{}) *Packer {
	if USE_BIGENDIAN {
		binary.Write(packer.out_, binary.BigEndian, i)
		return packer
	}
	binary.Write(packer.out_, binary.LittleEndian, i)
	return packer
}

func (packer *Packer) packInt64(value int64) *Packer {
	if value >= 0 {
		if value <= int64(MAX_7BIT) {
			packer.write(uint8(value) | MP_FIXNUM)
		} else if value <= int64(MAX_15BIT) {
			packer.write(MP_INT16).write(int16(value))
		} else if value <= int64(MAX_31BIT) {
			packer.write(MP_INT32).write(int32(value))
		} else {
			packer.write(MP_INT64).write(value)
		}
	} else {
		if value >= -(int64(MAX_5BIT) + 1) {
			packer.write(int8(int16(value) | int16(MP_NEGATIVE_FIXNUM)))
		} else if value >= -(int64(MAX_7BIT) + 1) {
			packer.write(MP_INT8).write(int8(value))
		} else if value >= -(int64(MAX_15BIT) + 1) {
			packer.write(MP_INT16).write(int16(value))
		} else if value >= -(int64(MAX_31BIT) + 1) {
			packer.write(MP_INT32).write(int32(value))
		} else {
			packer.write(MP_INT64).write(value)
		}
	}
	return packer
}

func (packer *Packer) packUInt64(value uint64) *Packer {
	if value <= uint64(MAX_7BIT) {
		packer.write(int8(value) | int8(MP_FIXNUM))
	} else if value <= uint64(MAX_8BIT) {
		packer.write(MP_UINT8).write(uint8(value))
	} else if value <= uint64(MAX_16BIT) {
		packer.write(MP_UINT16).write(uint16(value))
	} else if value <= uint64(MAX_32BIT) {
		packer.write(MP_UINT32).write(uint32(value))
	} else {
		packer.write(MP_UINT64).write(value)
	}
	return packer
}

func (packer *Packer) packBytes(value []byte) {
	length := uint32(len(value))
	if length <= MAX_5BIT {
		packer.write(int8(uint8(length) | MP_FIXRAW))
	} else if length <= MAX_16BIT {
		packer.write(MP_RAW16).write(int16(length))
	} else {
		packer.write(MP_RAW32).write(int32(length))
	}
	packer.out_.Write(value)
}

func (packer *Packer) GetBuffer() []byte {
	return packer.out_.Bytes()
}

func (packer *Packer) ClearBuffer() {
	packer.out_.Reset()
}

func (packer *Packer) PackFloat(value float32) {
	glog.LogConsole(glog.LogInfo, "PackFloat", value)
	packer.write(MP_FLOAT).write(math.Float32bits(value))
}

func (packer *Packer) PackDouble(value float64) {
	glog.LogConsole(glog.LogInfo, "PackDouble", value)
	packer.write(MP_DOUBLE).write(math.Float64bits(value))
}

func (packer *Packer) PackBool(value bool) {
	glog.LogConsole(glog.LogInfo, "PackBool", value)
	if value {
		packer.write(MP_TRUE)
		return
	}
	packer.write(MP_FALSE)
}

func (packer *Packer) PackBytes(value []byte) {
	glog.LogConsole(glog.LogInfo, "PackBytes", value)
	packer.packBytes(value)
}

func (packer *Packer) PackString(value string) {
	glog.LogConsole(glog.LogInfo, "PackString", value)
	packer.packBytes([]byte(value))
}

func (packer *Packer) PackInt32(value int32) {
	glog.LogConsole(glog.LogInfo, "PackInt32", value)
	packer.packInt64(int64(value))
}

func (packer *Packer) PackInt64(value int64) {
	glog.LogConsole(glog.LogInfo, "PackInt64", value)
	packer.packInt64(value)
}

func (packer *Packer) PackUInt32(value uint32) {
	glog.LogConsole(glog.LogInfo, "PackUInt32", value)
	packer.packUInt64(uint64(value))
}

func (packer *Packer) PackUInt64(value uint64) {
	glog.LogConsole(glog.LogInfo, "PackUInt64", value)
	packer.packUInt64(value)
}

type UnPacker struct {
	in_ *bytes.Buffer
}

func (unpacker *UnPacker) read(i interface{}) (err error) {
	if USE_BIGENDIAN {
		err = binary.Read(unpacker.in_, binary.BigEndian, i)
	}
	err = binary.Read(unpacker.in_, binary.LittleEndian, i)
	return
}

func (unpacker *UnPacker) readString(slen uint32) (int, string) {
	data := unpacker.in_.Next(int(slen))
	if len(data) == int(slen) {
		return 0, string(data)
	}
	return -1, ""
}

func (unpacker *UnPacker) unpack() (int, interface{}) {
	var header uint8
	if err := unpacker.read(&header); err != nil {
		return -1, nil
	}
	var (
		fVal      float32
		dVal      float64
		int8Val   int8
		int16Val  int16
		int32Val  int32
		int64Val  int64
		uint8Val  uint8
		uint16Val uint16
		uint32Val uint32
		uint64Val uint64
	)
	glog.LogConsole(glog.LogInfo, "unpack header", header)
	switch header {
	case MP_UINT8:
		if err := unpacker.read(&uint8Val); err != nil {
			return -1, 0
		}
		return 0, uint8Val
	case MP_UINT16:
		if err := unpacker.read(&uint16Val); err != nil {
			return -1, 0
		}
		return 0, uint16Val
	case MP_UINT32:
		if err := unpacker.read(&uint32Val); err != nil {
			return -1, 0
		}
		return 0, uint32Val
	case MP_UINT64:
		if err := unpacker.read(&uint64Val); err != nil {
			return -1, 0
		}
		return 0, uint64Val
	case MP_INT8:
		if err := unpacker.read(&int8Val); err != nil {
			return -1, 0
		}
		return 0, int8Val
	case MP_INT16:
		if err := unpacker.read(&int16Val); err != nil {
			return -1, 0
		}
		return 0, int16Val
	case MP_INT32:
		if err := unpacker.read(&int32Val); err != nil {
			return -1, 0
		}
		return 0, int32Val
	case MP_INT64:
		if err := unpacker.read(&int64Val); err != nil {
			return -1, 0
		}
		return 0, int64Val
	case MP_FLOAT:
		if err := unpacker.read(&fVal); err != nil {
			return -1, 0
		}
		return 0, fVal
	case MP_DOUBLE:
		if err := unpacker.read(&dVal); err != nil {
			return -1, 0
		}
		return 0, dVal
	case MP_NULL:
		return 0, nil
	case MP_FALSE:
		return 0, false
	case MP_TRUE:
		return 0, true
	case MP_ARRAY16:
		fallthrough
	case MP_ARRAY32:
		fallthrough
	case MP_MAP16:
		fallthrough
	case MP_MAP32:
		return -2, nil
	case MP_RAW16:
		if err := unpacker.read(&uint16Val); err != nil {
			return -1, nil
		}
		return unpacker.readString(uint32(uint16Val))
	case MP_RAW32:
		if err := unpacker.read(&uint32Val); err != nil {
			return -1, nil
		}
		return unpacker.readString(uint32(uint32Val))
	}

	if (header & uint8(0xE0)) == MP_FIXRAW {
		uint8Val = header - MP_FIXRAW
		return unpacker.readString(uint32(uint8Val))
	}

	if (header & uint8(0xE0)) == MP_NEGATIVE_FIXNUM {
		int8Val = int8(header&uint8(0x1F)) - 32
		return 0, int8Val
	}

	if (header & uint8(0xF0)) == MP_FIXARRAY {
		header = MP_FIXARRAY
		return -2, nil
	}

	if (header & uint8(0xF0)) == MP_FIXMAP {
		header = MP_FIXMAP
		return -2, nil
	}

	if header <= 127 {
		uint8Val = header
		return 0, uint8Val
	}

	return -4, nil
}

func (unpacker *UnPacker) Attatch(data []byte) {
	unpacker.in_.Reset()
	unpacker.in_.Write(data)
}

func (unpacker *UnPacker) UnPackUInt64() (r int, value uint64) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case uint8:
			r, value = 0, uint64(v)
		case uint16:
			r, value = 0, uint64(v)
		case uint32:
			r, value = 0, uint64(v)
		case uint64:
			r, value = 0, v
		default:
			r, value = -1, 0
		}
	} else {
		r, value = -1, 0
	}
	glog.LogConsole(glog.LogInfo, "UnPackUInt64", value)
	return
}

func (unpacker *UnPacker) UnPackUInt32() (r int, value uint32) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case uint8:
			r, value = 0, uint32(v)
		case uint16:
			r, value = 0, uint32(v)
		case uint32:
			r, value = 0, v
		default:
			r, value = -1, 0
		}
	} else {
		r, value = -1, 0
	}
	glog.LogConsole(glog.LogInfo, "UnPackUInt32", value)
	return
}

func (unpacker *UnPacker) UnPackInt32() (r int, value int32) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case uint8:
			r, value = 0, int32(v)
		case uint16:
			r, value = 0, int32(v)
		case int8:
			r, value = 0, int32(v)
		case int16:
			r, value = 0, int32(v)
		case int32:
			r, value = 0, v
		default:
			r, value = -1, 0
		}
	} else {
		r, value = -1, 0
	}
	glog.LogConsole(glog.LogInfo, "UnPackInt32", r, value)
	return
}

func (unpacker *UnPacker) UnPackInt64() (r int, value int64) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case uint8:
			r, value = 0, int64(v)
		case uint16:
			r, value = 0, int64(v)
		case uint32:
			r, value = 0, int64(v)
		case int8:
			r, value = 0, int64(v)
		case int16:
			r, value = 0, int64(v)
		case int32:
			r, value = 0, int64(v)
		case int64:
			r, value = 0, v
		default:
			r, value = -1, 0
		}
	} else {
		r, value = -1, 0
	}
	glog.LogConsole(glog.LogInfo, "UnPackInt64", r, value)
	return
}

func (unpacker *UnPacker) UnPackBytes() (r int, value []byte) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case string:
			r, value = _r, []byte(v)
		default:
			r, value = -1, nil
		}
	} else {
		r, value = -1, nil
	}
	glog.LogConsole(glog.LogInfo, "UnPackBytes", r, value)
	return
}

func (unpacker *UnPacker) UnPackString() (r int, value string) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case string:
			r, value = _r, string(v)
		default:
			r, value = -1, ""
		}
	} else {
		r, value = -1, ""
	}
	glog.LogConsole(glog.LogInfo, "UnPackString", r, value)
	return
}

func (unpacker *UnPacker) UnPackFloat() (r int, value float32) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case float32:
			r, value = _r, float32(v)
		default:
			r, value = -1, 0
		}
	} else {
		r, value = -1, 0
	}
	glog.LogConsole(glog.LogInfo, "UnPackFloat", r, value)
	return
}

func (unpacker *UnPacker) UnPackDouble() (r int, value float64) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case float64:
			r, value = _r, float64(v)
		default:
			r, value = -1, 0
		}
	} else {
		r, value = -1, 0
	}
	glog.LogConsole(glog.LogInfo, "UnPackDouble", r, value)
	return
}

func (unpacker *UnPacker) UnPackBool() (r int, value bool) {
	if _r, _value := unpacker.unpack(); _r == 0 {
		switch v := _value.(type) {
		case bool:
			r, value = _r, bool(v)
		default:
			r, value = -1, false
		}
	} else {
		r, value = -1, false
	}
	glog.LogConsole(glog.LogInfo, "UnPackBool", r, value)
	return
}
