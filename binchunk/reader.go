package binchunk

import (
	"encoding/binary"
	"math"
)

type reader struct {
	data []byte
}

// 读取基本数据类型
// readByte:读字节
func (self *reader) readByte() byte {
	b := self.data[0]
	self.data = self.data[1:]
	return b
}

// readUint32:读32位整数
func (self *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(self.data)
	self.data = self.data[4:]
	return i
}

// readUint64:读64位整数
func (self *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(self.data)
	self.data = self.data[8:]
	return i
}

// readLuaInteger:读一个Lua整数
func (self *reader) readLuaInteger() int64 {
	return int64(self.readUint64())
}

// readLuaNumber:读一个Lua浮点数Number类型
func (self *reader) readLuaNumber() float64 {
	return math.Float64frombits(self.readUint64())
}

// readString: 根据首个字节长度标识判断为长字符串还是短字符串并读取
func (self *reader) readString() string {
	// 读取第一个字节
	size := uint(self.readByte()) // 短字符串
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(self.readUint64()) // 长字符串
	}
	bytes := self.readBytes(size - 1)
	return string(bytes)
}

// readBytes: 从字节流中读取N个字节
func (self *reader) readBytes(n uint) []byte {
	bytes := self.data[:n]
	self.data = self.data[n:]
	return bytes
}

// checkHeader: 检查头部
func (self *reader) checkHeader() {
	if string(self.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
	}
	if self.readByte() != LUAC_VERSION {
		panic("version mismatch!")
	}
	if self.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
	}
	if string(self.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
	}
	if self.readByte() != CINT_SIZE {
		panic("int size mismatch!")
	}
	if self.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
	}
	if self.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
	}
	if self.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
	}
	if self.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
	}
	if self.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch!")
	}
	if self.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}

// readProto: 读取函数原型
func (self *reader) readProto(parentSource string) *Prototype {
	source := self.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     self.readUint32(),
		LastLineDefined: self.readUint32(),
		NumParams:       self.readByte(),
		IsVararg:        self.readByte(),
		MaxStackSize:    self.readByte(),
		Code:            self.readCode(),
		Constants:       self.readConstants(),
		Upvalues:        self.readUpvalues(),
		Protos:          self.readProtos(source),
		LineInfo:        self.readLineInfo(),
		LocVars:         self.readLocVars(),
		UpvalueNames:    self.readUpvalueNames(),
	}
}

// readCode: 读取指令表
func (self *reader) readCode() []uint32 {
	code := make([]uint32, self.readUint32())
	for i := range code {
		code[i] = self.readUint32()
	}
	return code
}

// readConstants: 读取常量表
func (self *reader) readConstants() []interface{} {
	constants := make([]interface{}, self.readUint32())
	for i := range constants {
		constants[i] = self.readConstant()
	}
	return constants
}

// readConstant: 读取常量
func (self *reader) readConstant() interface{} {
	switch self.readByte() {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return self.readByte() != 0
	case TAG_INTEGER:
		return self.readLuaInteger()
	case TAG_NUMBER:
		return self.readLuaNumber()
	case TAG_SHORT_STR, TAG_LONG_STR:
		return self.readString()
	default:
		panic("corrupted!") // todo
	}
}

// readUpvalues: 读取Upvalue表
func (self *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, self.readUint32())
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: self.readByte(),
			Idx:     self.readByte(),
		}
	}
	return upvalues
}

// readProtos: 读取内嵌函数原型
func (self *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, self.readUint32())
	for i := range protos {
		protos[i] = self.readProto(parentSource)
	}
	return protos
}

// readLineInfo: 读取行号表
func (self *reader) readLineInfo() []uint32 {
	lineInfo := make([]uint32, self.readUint32())
	for i := range lineInfo {
		lineInfo[i] = self.readUint32()
	}
	return lineInfo
}

// readLocVars: 读取局部变量表
func (self *reader) readLocVars() []LocVar {
	locVars := make([]LocVar, self.readUint32())
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: self.readString(),
			StartPC: self.readUint32(),
			EndPC:   self.readUint32(),
		}
	}
	return locVars
}

// readUpvalueNames: 读取Upvalue名字表
func (self *reader) readUpvalueNames() []string {
	names := make([]string, self.readUint32())
	for i := range names {
		names[i] = self.readString()
	}
	return names
}
