package main

import (
	"encoding/json"
	"fmt"
	. "luago/compiler/lexer"
	"luago/compiler/parser"
	"luago/state"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		ls := state.New()
		ls.OpenLibs()           // 开启标准库
		ls.LoadFile(os.Args[0]) // 加载文件
		ls.Call(0, -1)
	}

}

// 测试模块使用
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
