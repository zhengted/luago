package parser

import (
	. "luago/compiler/ast"
	. "luago/compiler/lexer"
)

func Parse(chunk, chunkName string) *Block {
	lexer := NewLexer(chunk, chunkName)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_EOF)
	return block
}
