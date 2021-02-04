

* [实现Lua虚拟机、编译器和标准库](#实现lua虚拟机编译器和标准库)
* [Table of Contents](#table-of-contents)
  * [Lua虚拟机和API](#lua虚拟机和api)
    * [二进制Chunk](#二进制chunk)
      * [整体结构](#整体结构)
      * [如何解析BinaryChunk](#如何解析binarychunk)
      * [检查头部和读取函数原型](#检查头部和读取函数原型)
    * [指令集](#指令集)
      * [编码模式](#编码模式)
      * [操作数](#操作数)
      * [指令解码](#指令解码)
      * [打印解码内容](#打印解码内容)
    * [API](#api)
      * [LuaAPI、LuaState和宿主程序的关系](#luaapiluastate和宿主程序的关系)
      * [关于Lua栈的索引计算](#关于lua栈的索引计算)
      * [LuaState](#luastate)
      * [X方法](#x方法)
    * [运算符](#运算符)
      * [Lua运算符介绍](#lua运算符介绍)
      * [自动类型转换](#自动类型转换)
    * [虚拟机雏形](#虚拟机雏形)
      * [PC（Programme Counter）](#pcprogramme-counter)
      * [指令封装](#指令封装)
      * [for循环](#for循环)
        * [forprep](#forprep)
        * [forloop](#forloop)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

实现Lua虚拟机、编译器和标准库
=================

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

- 指令解码的方式主要是位运算，理解这一部分的位运算对Lua虚拟机的底层内存结构印象会深一些

  ```go
  const (
  	MAXARG_Bx  = 1<<18 - 1      // 2^18 - 1 = 262143
  	MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071
  )
  
  type Instruction uint32
  
  // Opcode:提取操作码
  func (self Instruction) Opcode() int {
  	return int(self & 0x3F)
  }
  
  // ABC:ABC模式提取操作数
  func (self Instruction) ABC() (a, b, c int) {
  	a = int(self >> 6 & 0xFF)
  	b = int(self >> 14 & 0x1FF)
  	c = int(self >> 23 & 0x1FF)
  	return
  }
  
  // ABx: ABx提取操作数
  func (self Instruction) ABx() (a, bx int) {
  	a = int(self >> 6 & 0xFF)
  	bx = int(self >> 14)
  	return
  }
  
  // AsBx: AsBx提取操作数
  func (self Instruction) AsBx() (a, sbx int) {
  	a, bx := self.ABx()
  	return a, bx - MAXARG_sBx
  }
  
  // 以上两者的区别在于sbx是有符号的，而bx是无符号的
  // sbx的取值范围是：-131071~131072
  // bx的取值范围是：0~262143
  
  // Ax: Ax提取参数
  func (self Instruction) Ax() int {
  	return int(self >> 6)
  }
  ```

- 其中ABx和AsBx的解码需注意两者的范围是不同的

  ![image-20210127210045395](https://i.loli.net/2021/01/27/75UhyQweZ8W4q1i.png)

#### 打印解码内容



```go
func printOperands(i Instruction) {
	switch i.OpMode() {
	case IABC:
		a, b, c := i.ABC()

		fmt.Printf("%d", a)
		if i.BMode() != OpArgN {
			if b > 0xFF {
				// 最高比特位为1，常量表索引，按负数输出
				fmt.Printf(" %d", -1-b&0xFF)
			} else {
				// 最高比特位为0，寄存器索引，正常输出
				fmt.Printf(" %d", b)
			}
		}
		if i.CMode() != OpArgN {
			if c > 0xFF {
				//同上
				fmt.Printf(" %d", -1-c&0xFF)
			} else {
				fmt.Printf(" %d", c)
			}
		}
	case IABx:
		a, bx := i.ABx()
		fmt.Printf("%d", a)
        // 这里则是根据操作数类型决定打印出来的是有符号数还是无符号数
		if i.BMode() == OpArgK {
			fmt.Printf(" %d", -1-bx)
		} else if i.BMode() == OpArgU {
			fmt.Printf(" %d", bx)
		}
	case IAsBx:
		a, sbx := i.AsBx()
		fmt.Printf("%d %d", a, sbx)
	case IAx:
		ax := i.Ax()
		fmt.Printf("%d", -1-ax)
	}
}
```

### API

#### LuaAPI、LuaState和宿主程序的关系

![image-20210128115559475](https://i.loli.net/2021/01/28/MglPTkSiF5QRmAD.png)

#### 关于Lua栈的索引计算

- 绝对索引：从栈底从下往上数是第几个元素就是几，范围[1,n]

- 相对索引：栈顶为-1，从上往下递减

  ![image-20210128120042812](https://i.loli.net/2021/01/28/YH6ienIsBagF153.png)

- 索引校验函数

  ```go
  // absIndex: 把索引转换为绝对索引
  // TODO:需考虑索引是否有效
  func (self *luaStack) absIndex(idx int) int {
  	if idx >= 0 {
  		return idx
  	}
  	return idx + self.top + 1
  }
  ```

- 栈中的操作大多针对索引，因此对索引的快速转换要掌握

#### LuaState

- LuaState包含的基本与栈相关的函数 如下
  - 基础栈的操纵方法，Push、Pop、Rotate等等
  - 栈访问方法，IsBoolean、ToBoolean等等
  - 压栈方法，PushBoolean、PushNumber等等

#### X方法

- 栈访问方法中有三组X方法，分别是ToNumber和ToNumberX、ToString和ToStringX以及ToInteger和ToIntegerX。以下以ToString为例

  ```go
  func (self *luaState) ToString(idx int) string {
  	s, _ := self.ToStringX(idx)
  	return s
  }
  
  func (self *luaState) ToStringX(idx int) (string, bool) {
  	val := self.stack.get(idx)
  
  	switch x := val.(type) {
  	case string:
  		return x, true
  	case int64, float64:
  		s := fmt.Sprintf("%v", x) // todo
  		self.stack.set(idx, s)
  		return s, true
  	default:
  		return "", false
  	}
  }
  ```

- ToString只在意获取到的结果，不关心是否成功

### 运算符

#### Lua运算符介绍

- 算数运算符：+、-、*、/、//（整除）、%、^（乘方）
- 按位运算符：&、|、~（二元异或，一元按位取反，**本代码中用两个符号代替**）、<<、>>
- 比较运算符：==、>、<
- 逻辑运算符：and、or、not
- 长度运算符：len
- 字符串拼接运算符：..

#### 自动类型转换

- 除法和乘方运算符
  1. 如果操作数是整数，则提升为浮点数
  2. 如果操作数是字符串，且可以解析为浮点数，则解析为浮点数
  3. 进行浮点数运算，结果也是浮点数

- 其他算术运算符
  1. 全为整数，进行整数运算
  2. 否则将操作数转换为浮点数，同除法和乘方运算
  3. 然后进行浮点数运算

- 按位运算符
  1. 全为整数，无需转换
  2. 操作数为浮点数，但实际表示的是整数值如100.0，则转换为整数
  3. 操作数是字符串，且可以解析为整数值如“100”，则解析为整数
  4. 操作数是字符串，且可以解析为要求2中的浮点数，则->浮点数->整数
  5. 进行整数运算

- 字符串拼接
  1. 操作数是字符串，则无需转换
  2. 操作数是数字，则->字符串
  3. 拼接操作

**为保证文档内容整洁，剩余实现部分请参考代码，api_arith、api_compare、api_misc**

### 虚拟机雏形

#### PC（Programme Counter）

- PC是程序计数器，用来记录当前的指令。任务是不停取出指令执行指令

  ```
  loop {
  	1.计算PC
  	2.取出当前指令
  	3.执行当前指令
  }
  ```

- 为了实现PC，使用LuaVM接口，这个接口从LuaState控制，结构如下

  ```go
  type LuaVM interface {
  	LuaState
  	PC() int          // 返回当前PC（仅测试用）
  	AddPC(n int)      // 修改PC（用于实现跳转指令）
  	Fetch() uint32    // 取出当前指令；将PC指向下一条指令
  	GetConst(idx int) // 将指定常量推入栈顶
  	GetRK(rk int)     // 将指定常量或栈值推入栈顶
  }
  ```

#### 指令封装

- 如果根据binchunk中得到的指令进行执行会显得代码臃肿（大量的swich case）

- 方案是将部分指令封装，包括二元算术运算指令和按位运算指令

  - 增加一个通用二元运算的接口

    ```go
    /*************************** 运算符相关 **************************/
    // _binaryArith: R(A) := RK(B) op RK(C)
    func _binaryArith(i Instruction, vm LuaVM, op ArithOp) {
    	a, b, c := i.ABC()
    	a += 1
    	vm.GetRK(b) // 将指定（常量或寄存器索引的值）推入栈顶
    	vm.GetRK(c)
    	vm.Arith(op) // 二元运算并将结果赋给栈顶
    	vm.Replace(a)
    }
    ```

  - 所有的二元运算都使用以上接口，以运算sub为例

    ```go
    func sub(i Instruction, vm LuaVM)  { 
        _binaryArith(i, vm, LUA_OPSUB) 
    }  // -
    ```

  - opcode结构体新增成员

    ```go
    type opcode struct {
    	testFlag byte // operator is a test (next instruction must be a jump)
    	setAFlag byte // instruction set register A
    	argBMode byte // B arg mode
    	argCMode byte // C arg mode
    	opMode   byte // op mode
    	name     string
    	action   func(i Instruction, vm api.LuaVM)
    }
    ```

**其他的指令处理不赘述了，看源码和图吧，我累了QAQ**

- 二元运算符示例图

  ![image-20210129165738608](https://i.loli.net/2021/01/29/Fw8zV5EKZSr7GOb.png)

- 一元运算符示例图

  ![image-20210129165815064](https://i.loli.net/2021/01/29/6A1FIHdNbxom3TY.png)

- 比较指令示意图（不会修改栈状态）

  ![image-20210129170123818](https://i.loli.net/2021/01/29/YyuGkEWFJxlg42d.png)

#### for循环

- for循环难度比较大，指令内容不好理解，单独出一部分进行讲解

##### forprep

- 先看代码

  ```go
  // forPrep:R(A) -= R(A+2); pc += sBx
  // 循环开始前预先给数值减去步长，然后跳转到FORLOOP指令开始循环
  func forPrep(i Instruction, vm LuaVM) {
  	a, sBx := i.AsBx()
  	a += 1
  
  	if vm.Type(a) == LUA_TSTRING {
  		vm.PushNumber(vm.ToNumber(a))
  		vm.Replace(a)
  	}
  	if vm.Type(a+1) == LUA_TSTRING {
  		vm.PushNumber(vm.ToNumber(a + 1))
  		vm.Replace(a + 1)
  	}
  	if vm.Type(a+2) == LUA_TSTRING {
  		vm.PushNumber(vm.ToNumber(a + 2))
  		vm.Replace(a + 2)
  	}
  	// R(A) -= R(A+2)
  	vm.PushValue(a)
  	vm.PushValue(a + 2)
  	vm.Arith(LUA_OPSUB)
  	vm.Replace(a)
  	// pc += sBx
  	vm.AddPC(sBx)
  }
  ```

- 指令目的：初始化循环，让index先减去步长，让循环能够第一次开始

- 图

  ![image-20210129171957990](https://i.loli.net/2021/01/29/4gO9XWY5AqZPrNV.png)

##### forloop

- forloop的目的则是直接让数值加上步长，如果超出范围则循环结束，否则开始执行代码块

- 代码

  ```go
  // forLoop: R(A) += R(A+2) if R(A) <= R(A+1) then pc+=sBx; R(A+3) = R(A)
  //	和ForPrep不一样，先给数值加上步场，然后判断是否在范围内，再执行循环体内的代码
  func forLoop(i Instruction, vm LuaVM) {
  	a, sBx := i.AsBx()
  	a += 1
  
  	// R(A) += R(A+2)
  	vm.PushValue(a + 2)
  	vm.PushValue(a)
  	vm.Arith(LUA_OPADD)
  	vm.Replace(a)
  
  	// R(A) <?= R(A+1)
  	isPositiveStep := vm.ToNumber(a+2) >= 0
  	if isPositiveStep && vm.Compare(a, a+1, LUA_OPLE) ||
  		!isPositiveStep && vm.Compare(a+1, a, LUA_OPLE) {
  		vm.AddPC(sBx)
  		vm.Copy(a, a+3)
  	}
  }
  ```

- 图

  ![image-20210129172246634](https://i.loli.net/2021/01/29/rBhc5mp2jkUGAMV.png)

### 表

#### 结构

```go
type luaTable struct {
   arr  []luaValue            // lua表数组部分
   _map map[luaValue]luaValue // lua表哈希部分
}
```

#### Lua表是如何扩容的

- 思考：Lua表定义 local t = {1,2}，此时如果定义t[4] = "ok"，4的存储是在哈希表还是数组？如果再向表中插入键值为3的数据，lua表的存储结构会发生什么变化？先插入3和先插入4，表的变化情况是什么样的？

- 解答：定义使用的是setlist指令，此时的表是有序的。索引4已经超出了数组部分的长度，因此存在哈希部分中，如果再插入索引3，3刚好满足长度加一，会使数组部分扩容，同时扩容后会在哈希表中查找符合**条件**的递增索引，比如4、5、6（如果都有的话）插入到数组部分中。收缩同理。

- 代码如下：

  ```go
  // _shrinkArray: 删除数组中多余的hole（值为nil的key）
  func (self *luaTable) _shrinkArray() {
  	for i := len(self.arr) - 1; i >= 0; i-- {
  		if self.arr[i] == nil {
  			self.arr = self.arr[0:i]
  		}
  	}
  }
  
  // _expandArray: 数组动态扩展
  func (self *luaTable) _expandArray() {
  	for idx := int64(len(self.arr)) + 1; true; idx++ {
  		if val, found := self._map[idx]; found {
  			delete(self._map, idx)
  			self.arr = append(self.arr, val)
  		} else {
  			break
  		}
  	}
  }
  ```

#### 核心方法

- createTable&newTable 建表

  - 区别在于是否携带了初始化大小
  - 伪代码表示为 R(A) := {} (size = B,C)
  - ![image-20210201162328743](https://i.loli.net/2021/02/01/l3BxG1aOkU5JrHc.png)

- setTable  表内元素设值

  - 伪代码表示为 R(A)[Rk(B)] := RK(C)，图以寄存器为例，常量例不赘述
  - ![image-20210201162850075](https://i.loli.net/2021/02/01/HskWdAZ7pz2FMoO.png)

- getTable  取出某个键对应的值

  - 伪代码表示 R(A) := R(B)[RK(C)]
  - ![image-20210201163000507](https://i.loli.net/2021/02/01/CwXaLW1H3bPqGE5.png)

- （番外）setList  批量设值

  - 这个方法用于初始化表定义。如果表初始化方式是这样的

    ```lua
    local t = {1,2,3,4}
    ```

    则表中的四个常量会先加载到常量表，再加载到寄存器中。SETLIST通过获取寄存器中的值初始化lua表结构

    具体操作参考setList(i Instruction, vm LuaVM)方法

**核心方法**也是对应指令。有了前面的铺垫，这部分的开发内容比较简单，难点在于理解lua表结构和扩容机制

### 函数调用

#### Lua函数调用的机制

- 参数灵活

  - 声明是a个参数，可以传a+n个参数，只是后面n个不会被接收

  - 赋值返回同理

    ```lua
    function f()
        return 1,2,3
    end
    a,b = (f())		-- 这种情况 a = 1, b = nil
    a,b,c = 5,f(),5	-- a = 5,b = 1,c = 5
    ```

    

#### 调用栈

- 函数调用是借助**调用栈**实现的，函数调用栈中存放的是**调用栈帧**

##### 调用函数的过程

1. 先向调用栈中推入一个调用帧，并把参数传递给调用帧。

2. 函数依托调用帧执行指令，期间可能会调用其他函数。

3. 当函数执行完毕后，调用帧里会留下函数需要返回的值。

4. 将调用帧弹出，并且把返回值返回给底部的调用帧

   ![image-20210204102526075](https://i.loli.net/2021/02/04/9iTJjAkFCK2XqeD.png)

   - f()内部调用g()，g()内部调用h()

```go
// 调用栈结构
type luaStack struct {
	slots   []luaValue
	top     int
	prev    *luaStack  // prev:与函数执行没有关系，让调用帧变成链表结点，不要被字面意思误导
	closure *closure   // closure:闭包，可以理解成函数原型
	varargs []luaValue // varargs:变长参数，
	pc      int        // pc:指令计数器
}
```

- push和pop方法用于明确当前使用的是哪个调用帧

![image-20210204103226492](https://i.loli.net/2021/02/04/4AUpF1aImfE3XY8.png)

#### 函数调用API

- 解释器在执行脚本之前，先把脚本装进一个主函数，然后把主函数编译成proto交给lua虚拟机去运行

##### Load方法

- 把主函数原型实例化为闭包并推入栈顶，除了能加载预编译的chunk，也可以直接加载Lua脚本
- 示意图如下

![image-20210204104428743](https://i.loli.net/2021/02/04/hz3BAC1xLJvoTOd.png)

```go
/*
	Load: 加载chunk，可以是lua也可以是编译后的二进制chunk，根据mode来决定
		返回值为状态码：0表示成功
*/
func (self *luaState) Load(chunk []byte, chunkName, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	self.stack.push(c)
	return 0
}
```

##### Call方法

- 对Lua函数进行调用，先把被调函数和参数推入栈顶，Call方法结束后，参数和被调函数被替换成返回值。遵循**多退少补**的原则

- 代码如下

  ```go
  // Call: 函数调用
  func (self *luaState) Call(nArgs, nResults int) {
  	val := self.stack.get(-(nArgs + 1))
  	if c, ok := val.(*closure); ok {
  		fmt.Printf("call %s<%d,%d>\n", c.proto.Source, c.proto.LineDefined,
  			c.proto.LastLineDefined)
  		self.callLuaClosure(nArgs, nResults, c)
  	} else {
  		panic("not function")
  	}
  }
  ```

###### 详细过程

**可参考上面的[调用函数](#调用函数的过程)的过程**，这里只做代码展示

```go
// callLuaClosure:具体逻辑，
func (self *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	// 1. 初始化信息，确定寄存器的数量，定义函数时声明的固定参数数量、
	//	以及是否是vararg函数（会是当扩大）
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	// 2. 创建一个新的调用帧，把闭包（函数原型）和调用帧联系起来
	newStack := newLuaStack(nRegs + 20)
	newStack.closure = c

	// 3. 调用popN把函数和参数值一次性从栈顶弹出，
	//	然后调用新帧的pushN方法按照固定参数数量传入参数
	funcAndArgs := self.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	// 4. 将新帧push进调用栈栈顶，让他成为当前帧，最后调用runLuaClosure
	self.pushLuaStack(newStack)
	self.runLuaClosure()
	self.popLuaStack()

	// 5. 将结果压入旧的调用栈中
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		self.stack.push(len(results))
		self.stack.pushN(results, nResults)
	}
}
```

#### 函数调用指令

- closure：把当前Lua函数的子原型实例化为闭包。从当前闭包的子函数原型表中取出原型（用Bx去索引），实例化为闭包，并推入LuaState栈顶，再从栈顶弹出赋值到指定寄存器（A）上

  ```go
  func closure(i Instruction, vm LuaVM) {
  	a, bx := i.ABx()
  	a += 1
  	vm.LoadProto(bx)
  	vm.Replace(a)
  }
  ```

- call：调用Lua函数。A：被调函数在寄存器中的索引（记得加一），B：决定参数数量，紧挨着A，C：决定返回值数量

  - 示意图如下

  ![image-20210204110729337](https://i.loli.net/2021/02/04/HWXtl1Ng7AC2SkI.png)

  - 代码

    ```go
    // call:iABC,R(A), ... , R(A+C-2) := R(A)(R(A+1),...,R(A+B-1))
    func call(i Instruction, vm LuaVM) {
    	a, b, c := i.ABC()
    	a += 1
    	nArgs := _pushFuncAndArgs(a, b, vm)	// 把被调函数和参数值推入栈顶
    	vm.Call(nArgs, c-1)
    	_popResult(a, c, vm)		//将原来栈顶的内容替换成指定数量的返回值
    }
    
    ```

  **这一段比较难，有机会写篇文档专门讲这块**

  - 关于_pushFuncAndArgs和\_popResult方法的解释
    - \_popResult方法：
      - 如果c > 1，按序replace栈即可。
      - 如果c == 1，返回值数量为0，无需操作。
      - 如果c == 0，需要把被调函数的返回值全部返回，并往栈顶推入一个整数值标记这些返回值原本要移动到哪些寄存器中。这种情况其实是编译器不知道调用之后会有几个返回值，即参数中有一个被调函数，如这种```f(1,2,g())```
      - 关于第三种情况其实可以不需要操作，因为在下次使用时一定是在栈顶，只需要入栈一个返回值初始位置索引
    - _pushFuncAndArgs和上面一样
      - 如果b >= 1，按序入栈即可，返回b - 1表示参数个数
      - b = 0，参数数量不确定。先把栈顶取出，得到变长参数所在的寄存器的位置，再将函数和参数入栈。进行旋转让变长参数长度和变长参数到达栈顶

- return指令：把存放在连续多个寄存器里的值返回给主调函数。A决定第一个寄存器索引，寄存器数量由B决定
  - b = 1 无返回值 无需操作
  - b > 1 按序入栈即可
  - b = 0 特殊情况调用fixstack

- vararg指令：把传递给当前函数的变长参数加载到连续多个寄存器中。第一个寄存器由A指定，寄存器数量由B指定
  - b > 1 把b-1个vararg参数复制到寄存器
  - b == 0 把全部vararg参数复制到寄存器

- tailcall指令：用于解决调用栈溢出问题使被调函数重用主调函数的调用帧

- self指令：语法糖 obj: f(a,b,c) 等价于 obj.f(obj,a,b,c)

  - 把对象和方法拷贝到相邻的两个目标寄存器中，对象本身在寄存器中，索引由操作数B决定。方法名在常量表中，索引由操作数决定

    ![image-20210204154050509](https://i.loli.net/2021/02/04/jm5sKpIg6YqhWkE.png)