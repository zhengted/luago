package parser

import (
	. "luago/compiler/ast"
	. "luago/compiler/lexer"
)

func parsePrefixExp(lexer *Lexer) Exp {
	var exp Exp
	if lexer.LookAhead() == TOKEN_IDENTIFIER {
		line, name := lexer.NextIdentifier()
		exp = &NameExp{line, name}
	} else {
		exp = parseParensExp(lexer)
	}
	return _finishPrefixExp(lexer, exp)
}

func _finishPrefixExp(lexer *Lexer, exp Exp) Exp {
	for {
		switch lexer.LookAhead() {
		case TOKEN_SEP_LBRACK: // 用[] 访问对象
			lexer.NextToken()
			keyExp := parseExp(lexer)
			lexer.NextTokenOfKind(TOKEN_SEP_RBRACK)
			exp = &TableAccessExp{lexer.Line(), exp, keyExp}
		case TOKEN_SEP_DOT: // 用.访问对象
			lexer.NextToken()
			line, name := lexer.NextIdentifier()
			keyExp := &StringExp{line, name}
			exp = &TableAccessExp{line, exp, keyExp}
		case TOKEN_SEP_COLON, TOKEN_SEP_LPAREN, TOKEN_SEP_LCURLY, TOKEN_STRING:
			exp = _finishFuncCallExp(lexer, exp) //[`:` Name] args
		default:
			return exp
		}
	}
	//return exp
}

// 圆括号表达式
func parseParensExp(lexer *Lexer) Exp {
	lexer.NextTokenOfKind(TOKEN_SEP_LPAREN)
	exp := parseExp(lexer)
	lexer.NextTokenOfKind(TOKEN_SEP_RPAREN)

	switch exp.(type) {
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp:
		return &ParensExp{exp}
	}
	return exp
}

// 函数调用表达式
func _finishFuncCallExp(lexer *Lexer, prefixExp Exp) *FuncCallExp {
	nameExp := _parseNameExp(lexer) // [`:` Name]
	line := lexer.Line()
	args := _parseArgs(lexer) // args
	lastLine := lexer.Line()
	return &FuncCallExp{
		Line:      line,
		LastLine:  lastLine,
		PrefixExp: prefixExp,
		NameExp:   nameExp,
		Args:      args,
	}
}

func _parseNameExp(lexer *Lexer) *StringExp {
	if lexer.LookAhead() == TOKEN_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		return &StringExp{
			Line: line,
			Str:  name,
		}
	}
	return nil
}

func _parseArgs(lexer *Lexer) (args []Exp) {
	switch lexer.LookAhead() {
	case TOKEN_SEP_LPAREN: // `(` [explist] `)`
		lexer.NextToken()
		if lexer.LookAhead() != TOKEN_SEP_RPAREN {
			args = parseExpList(lexer)
		}
	case TOKEN_SEP_LCURLY: // `{` [explist] `}`
		args = []Exp{parseTableConstructorExp(lexer)}
	default: // literal string
		line, str := lexer.NextTokenOfKind(TOKEN_STRING)
		args = []Exp{&StringExp{line, str}}
	}
	return
}
