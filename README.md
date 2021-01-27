[TOC]

# 实现Lua虚拟机、编译器和标准库

## Lua虚拟机和API

### 二进制Chunk

#### 整体结构

- binary_Chunk的整体结构如下

  ```go
  type binaryChunk struct {
  	header							// 头部信息，加载时用于校验版本号，大小端格式
  	sizeUpvalues byte				// upvalue的大小
  	mainFunc     *Prototype			// 函数原型
  }
  ```

- header

  - 头部总共占用约30个字节，具体的内容如下：

    ```go
    type header struct {
    	// signature:签名。二进制文件的固定魔数，Lua二进制chunk的固定魔数是0x1B4C7561，写成Go语言字符串字面量为\x1bLua
    	signature [4]byte
    	// version:版本号。用于虚拟机加载二进制Chunk时的检查，计算方式是大版本号*16+小版本号（不考虑发布号）
    	version byte
    	// format:占用一个字节，官方使用的格式号为0
    	format byte
    	// luacData:占用六个字节，前两个字节为0x1993。后续四个字节分别为回车（0x0D）换行（0x0A）替换（0x1A）和另一个换行符
        // 注意：这个地方根据Lua版本不同会有一定差异，例如Lua5.1这里的信息是大小端，但是Lua5.3是根据luacData的分布确定大小端
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
    ```

- sizeUpvalues：记录了upvalue的个数

- mainFunc

  - 函数原型主要包含函数基本信息、指令表、常量表、upvalue表、子函数原型表以及调试信息；基本信息又包括源文件名、起止行号、固定参数 个数、是否是vararg函数以及运行函数所必要的寄存器数量；调试信息又包括行号表、局部变量表以及upvalue名列表。

    ```go
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
        StartPC uint32			// TODO:这两个变量暂时不知道什么意思
    	EndPC   uint32			//
    }
    ```

#### 如何解析BinaryChunk

- 使用encoding/binary包，读取字节流。Lua反编译出来的Luac文件有个特点，其中的部分二进制结构的头几个字节会告诉你接下来需要读取几个字节，比如读取指令表

  ```go
  // readCode: 读取指令表
  func (self *reader) readCode() []uint32 {
  	code := make([]uint32, self.readUint32())
  	for i := range code {
  		code[i] = self.readUint32()
  	}
  	return code
  }
  ```

- 读取字节流的常用方法如下，大部分的读取方法都是依据以下基本类型编写的

  ```go
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
  ```

#### 检查头部和读取函数原型

- binChunk的头部大部分为配置中的常量，如果加载的lua脚本编译出来的Luac出现与配置不一致会有提示

  ```go
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
  ```

- 函数原型读取只做代码展示，实现细节参考代码

  ```go
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
  ```

### 指令集

#### 编码模式

- 指令集的编码模式有四种，分别为：iABC、iABx、iAsBx、iAx
  - iABC：三个操作数，分别占用8、9、9个比特，共有39条指令
  - iABx：两个操作数（A，Bx），分别占用8、18个比特，共有3条指令
  - iAsBx：两个操作数（A，sBx），分别占用8、18个比特，共有4条指令
  - iAx：仅一个操作数，占用全部26个比特，共一条指令

- 码

  ```go
  // OpMode 操作模式
  const (
  	IABC = iota
  	IABx
  	IAsBx
  	IAx
  )
  ```

- 图

  ![image-20210127200208380](https://i.loli.net/2021/01/27/9MliXtVwGHKx18d.png)

#### 操作数

- 操作数A主要用来标识目标寄存器的索引，其他操作数按照其表示的信息可以分为四种类型：OpArgN、 OpArgU、 OpArgR、 OpArgK
  - OpArgN：不表示任何信息，不被使用的。比如MOVE指令，只操作A和B（iABC模式下），C操作数则是OpArgN类型

  - OpArgU：正常的被使用的操作数

  - OpArgR：在iABC下标识寄存器索引。在iAsBx下标识跳转偏移，比如该模式下的MOVE指令则可以用伪代码表示为R(A) := R(B)。其中A表示dst寄存器索引，B表示src寄存器索引

  - OpArgK：表示常量表索引或者寄存器索引。例子不赘述

    - 如何区分什么时候用常量表，什么时候用寄存器。依赖于B、C操作数的第一个比特，为1则表示常量表索引，为0则表示寄存器索引

  - 码

    ```go
    /* OpArgMask */
    const (
    	OpArgN = iota // argument is not used
    	OpArgU        // argument is used
    	OpArgR        // argument is a register or a jump offset
    	OpArgK        // argument is a constant or register/constant
    )
    ```

#### 指令解码

