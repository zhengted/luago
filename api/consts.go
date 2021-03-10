package api

const (
	LUA_TNONE = iota - 1 // None指无效索引的返回值类型
	LUA_TNIL
	LUA_TBOOLEAN
	LUA_TLIGHTUSERDATA
	LUA_TNUMBER
	LUA_TSTRING
	LUA_TTABLE
	LUA_TFUNCTION
	LUA_TUSERDATA
	LUA_TTHREAD
)

// 运算符和按位运算的常量符号
const (
	LUA_OPADD  = iota // +
	LUA_OPSUB         // -
	LUA_OPMUL         // *
	LUA_OPMOD         // %
	LUA_OPPOW         // ^
	LUA_OPDIV         // /
	LUA_OPIDIV        // //
	LUA_OPBAND        // &
	LUA_OPBOR         // |
	LUA_OPBXOR        // ~
	LUA_OPSHL         // <<
	LUA_OPSHR         // >>
	LUA_OPUNM         // -
	LUA_OPBNOT        // ~
)

// 比较运算常量
const (
	LUA_OPEQ = iota // ==
	LUA_OPLT        // <
	LUA_OPLE        // <=
)

const (
	LUA_MINSTACK            = 20                    // 最小栈大小
	LUAI_MAXSTACK           = 1000000               // lua栈的最大索引
	LUA_REGISTRYINDEX       = -LUAI_MAXSTACK - 1000 // 注册表的伪索引	luastate在操作时用这个值作为索引
	LUA_RIDX_GLOBALS  int64 = 2                     // 定义全局环境在注册表中的索引
	LUA_MULTRET             = -1

	LUA_MAXINTEGER = 1<<63 - 1
	LUA_MININTEGER = -1 << 63
)

// 错误处理相关
const (
	LUA_OK = iota
	LUA_YIELD
	LUA_ERRRUN
	LUA_ERRSYNTAX
	LUA_ERRMEM
	LUA_ERRGCMM
	LUA_ERRERR
	LUA_ERRFILE
)
