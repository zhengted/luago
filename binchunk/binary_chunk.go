package binchunk

const (
	// 以下为头部信息的常量
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type binaryChunk struct {
	header
	sizeUpvalues byte
	mainFunc     *Prototype
}

type header struct {
	// signature:签名。二进制文件的固定魔数，Lua二进制chunk的固定魔数是0x1B4C7561，写成Go语言字符串字面量为\x1bLua
	signature [4]byte
	// version:版本号。用于虚拟机加载二进制Chunk时的检查，计算方式是大版本号*16+小版本号（不考虑发布号）
	version byte
	// format:占用一个字节，官方使用的格式号为0
	format byte
	// luacData:占用六个字节，前两个字节为0x1993。后续四个字节分别为回车（0x0D）换行（0x0A）替换（0x1A）和另一个换行符
	luacData byte
	// 接下来5个字节分别记录cint、size_t、Lua虚拟机指令、lua整数和lua浮点数这五种数据结构在chunk中占用的字节数
	cintSize        byte // cintSize: 4 标识占用4个字节
	sizetSize       byte // sizetSize: 8 标识占用8个字节
	instructionSize byte // instructionSize: 4 标识占用4个字节
	luaIntegerSize  byte // luaIntegerSize: 8 标识占用8个字节
	luaNumberSize   byte // luaNumberSize: 8 标识占用8个字节

	luacInt int64   // luacInt:存放整数0x5678 8个字节
	luacNum float64 // luacNum:存放浮点数370.5 8个字节
}

type Prototype struct {
	// Source:记录源文件名，只有在主函数原型里该字段才有值 第一个字节表示文件名长度+1，后续为@文件名的字节流
	Source string
	// 记录起止行号，主函数中都是0，普通函数都需要大于0
	LineDefined     uint32 // LineDefined:起始行号
	LastLineDefined uint32 // LastLineDefined: 终止行号
	// NumParams: 固定参数个数
	NumParams byte
	// IsVararg: 是否有变长参数 有则为1 无则为0
	IsVararg byte
	// MaxStackSize: 寄存器数量，编译时计算
	MaxStackSize byte
	// Code: 指令表，每条指令占4个字节
	Code []uint32
	// Constants: 常量表，用于存放Lua代码里出现的字面量，以1字节的tag开头。tag参考↑
	Constants []interface{}
	// Upvalues: 用于存放Upvalue,一个元素占2字节
	Upvalues []Upvalue
	// Protos: 存放内部定义的子函数的原型
	Protos []*Prototype
	// LineInfo: 记录每条指令占用的行数
	LineInfo []uint32
	// LocVars: 局部变量表，记录局部变量
	LocVars []LocVar
	// UpvalueNames: Upvalue名列表
	UpvalueNames []string
}

type Upvalue struct {
	Instack byte
	Idx     byte
}

type LocVar struct {
	VarName string
	StartPC uint32
	EndPC   uint32
}

// Undump:用于解析二进制chunk
func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()
	reader.readByte()
	return reader.readProto("")
}

func IsBinaryChunk(data []byte) bool {
	return len(data) > 4 && string(data[:4]) == LUA_SIGNATURE
}
