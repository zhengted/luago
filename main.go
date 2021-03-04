package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	. "luago/api"
	. "luago/compiler/lexer"
	"luago/compiler/parser"
)

func main() {
	// Global Test demo

	// LUA VM TEST
	//data, err := ioutil.ReadFile("luac.out")
	//if err != nil {
	//	panic(err)
	//}
	//ls := state.New()
	//ls.Register("print", print)
	//ls.Register("getmetatable", getMetatable)
	//ls.Register("setmetatable", setMetatable)
	//ls.Register("pairs", pairs)
	//ls.Register("ipairs", iPairs)
	//ls.Register("error", error)
	//ls.Register("pcall", pCall)
	//
	//ls.Load(data, "luac.out", "b")
	//ls.Call(0, 0)

	// LUA LEX TEST
	//if len(os.Args) > 1 {
	data, err := ioutil.ReadFile("test.lua")
	if err != nil {
		panic(err)
	}

	//testLexer(string(data), "test.lua")
	testParser(string(data), "test.lua")
	//}
}

// 用来注册的go函数
func print(ls LuaState) int {
	nArgs := ls.GetTop()
	for i := 1; i <= nArgs; i++ {
		if ls.IsBoolean(i) {
			fmt.Printf("%t", ls.ToBoolean(i))
		} else if ls.IsString(i) {
			fmt.Print(ls.ToString(i))
		} else {
			fmt.Print(ls.TypeName(ls.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}

func getMetatable(ls LuaState) int {
	if !ls.GetMetatable(1) {
		ls.PushNil()
	}
	return 1
}

func setMetatable(ls LuaState) int {
	ls.SetMetatable(1)
	return 1
}

func next(ls LuaState) int {
	ls.SetTop(2)
	if ls.Next(1) {
		return 2
	}
	ls.PushNil()
	return 1
}

func pairs(ls LuaState) int {
	ls.PushGoFunction(next)
	ls.PushValue(1)
	ls.PushNil()
	return 3
}

func iPairs(ls LuaState) int {
	ls.PushGoFunction(_iPairsAux)
	ls.PushValue(1)
	ls.PushInteger(0)
	return 3
}

func _iPairsAux(ls LuaState) int {
	i := ls.ToInteger(2) + 1
	ls.PushInteger(i)
	if ls.GetI(1, i) == LUA_TNIL {
		return 1
	} else {
		return 2
	}
}

func error(ls LuaState) int {
	return ls.Error()
}

func pCall(ls LuaState) int {
	nArgs := ls.GetTop() - 1
	status := ls.PCall(nArgs, -1, 0)
	ls.PushBoolean(status == LUA_OK)
	ls.Insert(1)
	return ls.GetTop()
}

func testLexer(chunk, chunkName string) {
	lexer := NewLexer(chunk, chunkName)
	for {
		line, kind, token := lexer.NextToken()
		fmt.Printf("[%2d] [%-10s] %s\n",
			line, kindToCategory(kind), token)
		if kind == TOKEN_EOF {
			break
		}
	}
}

func testParser(chunk, chunkName string) {
	ast := parser.Parse(chunk, chunkName)
	b, err := json.Marshal(ast)
	if err != nil {
		panic(err)
	}
	println(string(b))
}

func kindToCategory(kind int) string {
	switch {
	case kind < TOKEN_SEP_SEMI:
		return "other"
	case kind <= TOKEN_SEP_RCURLY:
		return "separator"
	case kind <= TOKEN_OP_NOT:
		return "operator"
	case kind <= TOKEN_KW_WHILE:
		return "keyword"
	case kind == TOKEN_IDENTIFIER:
		return "identifier"
	case kind == TOKEN_NUMBER:
		return "number"
	case kind == TOKEN_STRING:
		return "string"
	default:
		return "other"
	}
}
