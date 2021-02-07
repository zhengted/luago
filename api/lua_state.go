package api

type LuaType = int
type ArithOp = int
type CompareOp = int
type GoFunction func(state LuaState) int // Go函数类型，参数是LuaState，返回值是Go函数返回值个数

// LuaUpvalueIndex:获取luaUpValue的伪索引
func LuaUpvalueIndex(i int) int {
	return LUA_REGISTRYINDEX - i
}

type LuaState interface {
	// 基本栈操作
	GetTop() int
	AbsIndex(idx int) int
	CheckStack(n int) bool
	Pop(n int)
	Copy(fromIdx, toIdx int)
	PushValue(idx int)
	Replace(idx int)
	Insert(idx int)
	Remove(idx int)
	Rotate(idx, n int)
	SetTop(idx int)
	// luaStack 访问 Go
	TypeName(tp LuaType) string
	Type(idx int) LuaType
	IsNone(idx int) bool
	IsNil(idx int) bool
	IsNoneOrNil(idx int) bool
	IsBoolean(idx int) bool
	IsInteger(idx int) bool
	IsNumber(idx int) bool
	IsString(idx int) bool
	IsTable(idx int) bool
	IsThread(idx int) bool
	IsFunction(idx int) bool
	ToBoolean(idx int) bool
	ToInteger(idx int) int64
	ToIntegerX(idx int) (int64, bool)
	ToNumber(idx int) float64
	ToNumberX(idx int) (float64, bool)
	ToString(idx int) string
	ToStringX(idx int) (string, bool)
	// Golang 访问 luaStack栈
	PushNil()
	PushBoolean(b bool)
	PushInteger(n int64)
	PushNumber(n float64)
	PushString(s string)

	// 运算符相关
	Arith(op ArithOp)                          // 执行算术和按位运算
	Compare(idx1, idx2 int, op CompareOp) bool // 比较运算
	Len(idx int)                               // 取长度计算
	Concat(n int)                              // 字符串拼接计算

	// 测试用
	PrintStack() // 打印栈

	/* get functions (Lua -> stack) 都是放到栈里的*/
	NewTable()
	CreateTable(nArr, nRec int)
	GetTable(idx int) LuaType
	GetField(idx int, k string) LuaType
	GetI(idx int, i int64) LuaType

	/* set functions (stack->Lua) */
	SetTable(idx int)
	SetField(idx int, k string)
	SetI(idx int, n int64)

	// 函数相关方法
	Load(chunk []byte, chunkName, mode string) int // 加载chunk（二进制chunk或Lua文件）
	Call(nArgs, nResult int)

	// Go调用相关方法
	PushGoFunction(f GoFunction)
	IsGoFunction(idx int) bool
	ToGoFunction(idx int) GoFunction

	// 全局表操作方法
	PushGlobalTable()
	GetGlobal(name string) LuaType
	SetGlobal(name string)
	Register(name string, f GoFunction)

	// 闭包相关
	PushGoClosure(f GoFunction, n int)
}
