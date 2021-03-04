package parser

import . "luago/compiler/ast"
import . "luago/compiler/lexer"

var _statEmpty = &EmptyStat{}

/*
stat ::=  ‘;’
	| break
	| ‘::’ Name ‘::’
	| goto Name
	| do block end
	| while exp do block end
	| repeat block until exp
	| if exp then block {elseif exp then block} [else block] end
	| for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	| for namelist in explist do block end
	| function funcname funcbody
	| local function Name funcbody
	| local namelist [‘=’ explist]
	| varlist ‘=’ explist
	| functioncall
*/
func parseStat(lexer *Lexer) Stat {
	switch lexer.LookAhead() {
	case TOKEN_SEP_SEMI:
		return parseEmptyStat(lexer)
	case TOKEN_KW_BREAK:
		return parseBreakStat(lexer)
	case TOKEN_SEP_LABEL:
		return parseLabelStat(lexer)
	case TOKEN_KW_GOTO:
		return parseGotoStat(lexer)
	case TOKEN_KW_DO:
		return parseDoStat(lexer)
	case TOKEN_KW_WHILE:
		return parseWhileStat(lexer)
	case TOKEN_KW_REPEAT:
		return parseRepeatStat(lexer)
	case TOKEN_KW_IF:
		return parseIfStat(lexer)
	case TOKEN_KW_FOR:
		return parseForStat(lexer)
	case TOKEN_KW_FUNCTION:
		return parseFuncDefStat(lexer)
	case TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(lexer)
	default:
		return parseAssignOrFuncCallStat(lexer)
	}
}

func parseEmptyStat(lexer *Lexer) *EmptyStat {
	lexer.NextTokenOfKind(TOKEN_SEP_SEMI)
	return _statEmpty
}

func parseBreakStat(lexer *Lexer) *BreakStat {
	lexer.NextTokenOfKind(TOKEN_KW_BREAK)
	return &BreakStat{Line: lexer.Line()}
}

func parseLabelStat(lexer *Lexer) *LabelStat {
	// 跳过分隔符记录签名
	lexer.NextTokenOfKind(TOKEN_SEP_LABEL)
	_, name := lexer.NextIdentifier()
	lexer.NextTokenOfKind(TOKEN_SEP_LABEL)
	return &LabelStat{Name: name}
}

func parseGotoStat(lexer *Lexer) *GotoStat {
	lexer.NextTokenOfKind(TOKEN_KW_GOTO)
	_, name := lexer.NextIdentifier()
	return &GotoStat{Name: name}
}

func parseDoStat(lexer *Lexer) *DoStat {
	lexer.NextTokenOfKind(TOKEN_KW_DO)  // 先跳过do
	block := parseBlock(lexer)          // 解析块
	lexer.NextTokenOfKind(TOKEN_KW_END) // 跳过END
	return &DoStat{Block: block}
}

func parseWhileStat(lexer *Lexer) *WhileStat {
	lexer.NextTokenOfKind(TOKEN_KW_WHILE) // while
	exp := parseExp(lexer)                // exp
	lexer.NextTokenOfKind(TOKEN_KW_DO)    // do
	block := parseBlock(lexer)            // block
	lexer.NextTokenOfKind(TOKEN_KW_END)   // end
	return &WhileStat{
		Exp:   exp,
		Block: block,
	}
}

func parseRepeatStat(lexer *Lexer) *RepeatStat {
	lexer.NextTokenOfKind(TOKEN_KW_REPEAT)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_UNTIL)
	exp := parseExp(lexer)
	return &RepeatStat{
		Block: block,
		Exp:   exp,
	}
}

func parseIfStat(lexer *Lexer) *IfStat {
	exps := make([]Exp, 0, 4)
	Blocks := make([]*Block, 0, 4)

	lexer.NextTokenOfKind(TOKEN_KW_IF)         // if
	exps = append(exps, parseExp(lexer))       // exp
	lexer.NextTokenOfKind(TOKEN_KW_THEN)       // then
	Blocks = append(Blocks, parseBlock(lexer)) // block

	for lexer.LookAhead() == TOKEN_KW_ELSEIF { // 判断是不是elseif
		lexer.NextToken()                          // elseif
		exps = append(exps, parseExp(lexer))       // exp
		lexer.NextTokenOfKind(TOKEN_KW_THEN)       // then
		Blocks = append(Blocks, parseBlock(lexer)) // block
	}

	if lexer.LookAhead() == TOKEN_KW_ELSE {
		lexer.NextToken()                           // else
		exps = append(exps, &TrueExp{lexer.Line()}) // 这里要存一个全真的表达式
		Blocks = append(Blocks, parseBlock(lexer))  // block
	}
	lexer.NextTokenOfKind(TOKEN_KW_END)
	return &IfStat{
		Exps:   exps,
		Blocks: Blocks,
	}
}

// For循环stat实现
func parseForStat(lexer *Lexer) Stat {
	lineOfFor, _ := lexer.NextTokenOfKind(TOKEN_KW_FOR)
	_, name := lexer.NextIdentifier()
	if lexer.LookAhead() == TOKEN_OP_ASSIGN {
		// 赋值语句 数值循环
		return _finishForNumStat(lexer, lineOfFor, name)
	} else {
		// 通用循环
		return _finishForInStat(lexer, name)
	}
}

// _finishForNumStat：数值循环
// lexer:扫描器 lineOfFor:for循环起始行号 name:关键字名
func _finishForNumStat(lexer *Lexer, lineOfFor int, name string) *ForNumStat {
	// 条件处理 初始条件和限制条件
	lexer.NextTokenOfKind(TOKEN_OP_ASSIGN) // for name `=`
	initExp := parseExp(lexer)             // exp
	lexer.NextTokenOfKind(TOKEN_SEP_COMMA) // `,`
	limitExp := parseExp(lexer)            // exp

	// 步长处理
	var stepExp Exp
	if lexer.LookAhead() == TOKEN_SEP_COMMA { // [
		lexer.NextToken()         // `,`
		stepExp = parseExp(lexer) // exp
	} else { // ]
		// 无定义用1
		stepExp = &IntegerExp{lexer.Line(), 1}
	}

	lineOfDo, _ := lexer.NextTokenOfKind(TOKEN_KW_DO) // do
	block := parseBlock(lexer)                        // block
	lexer.NextTokenOfKind(TOKEN_KW_END)               // end
	return &ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineOfDo,
		VarName:   name,
		InitExp:   initExp,
		LimitExp:  limitExp,
		StepExp:   stepExp,
		Block:     block,
	}
}

// _finishForInStat:泛型循环
func _finishForInStat(lexer *Lexer, name string) *ForInStat {
	nameList := _finishNameList(lexer, name)
	lexer.NextTokenOfKind(TOKEN_KW_IN)
	expList := parseExpList(lexer) // 常用的就是pairs和ipairs了

	lineOfDo, _ := lexer.NextTokenOfKind(TOKEN_KW_DO)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_END)

	return &ForInStat{
		LineOfDo: lineOfDo,
		NameList: nameList,
		ExpList:  expList,
		Block:    block,
	}
}

// _finishNameList: 变量名表
func _finishNameList(lexer *Lexer, name string) []string {
	names := []string{name}
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		_, nextName := lexer.NextIdentifier()
		names = append(names, nextName)
	}
	return names
}

func parseAssignOrFuncCallStat(lexer *Lexer) Stat {
	prefixExp := parsePrefixExp(lexer)
	if fc, ok := prefixExp.(*FuncCallExp); ok {
		return fc
	} else {
		return parseAssignStat(lexer, prefixExp)
	}
}

func parseAssignStat(lexer *Lexer, var0 Exp) *AssignStat {
	varList := _finishVarList(lexer, var0)
	lexer.NextTokenOfKind(TOKEN_OP_ASSIGN)
	expList := parseExpList(lexer)
	lastLine := lexer.Line()
	return &AssignStat{
		LastLine: lastLine,
		VarList:  varList,
		ExpList:  expList,
	}
}

func _finishVarList(lexer *Lexer, var0 Exp) []Exp {
	vars := []Exp{_checkVar(lexer, var0)}
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		exp := parsePrefixExp(lexer)
		vars = append(vars, _checkVar(lexer, exp))
	}
	return vars
}

func _checkVar(lexer *Lexer, exp Exp) Exp {
	switch exp.(type) {
	case *NameExp, *TableAccessExp:
		return exp
	}
	lexer.NextTokenOfKind(-1)
	panic("unreachable!")
}

// function funcname funcbody
// funcname ::= Name {‘.’ Name} [‘:’ Name]
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
func parseFuncDefStat(lexer *Lexer) *AssignStat {
	lexer.NextTokenOfKind(TOKEN_KW_FUNCTION) // function
	fnExp, hasColon := _parseFuncName(lexer) // funcName
	fdExp := parseFuncDefExp(lexer)          // funcBody
	if hasColon {                            // v:name(args) => v.name(self,args) Lua语法糖 传self
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
	}
	return &AssignStat{
		LastLine: fdExp.Line,
		VarList:  []Exp{fnExp},
		ExpList:  []Exp{fdExp},
	}

}

// local function Name funcbody
// local namelist [‘=’ explist]
func parseLocalAssignOrFuncDefStat(lexer *Lexer) Stat {
	lexer.NextTokenOfKind(TOKEN_KW_LOCAL)
	if lexer.LookAhead() == TOKEN_KW_FUNCTION {
		return _finishLocalFuncDefStat(lexer)
	} else {
		return _finishLocalVarDeclStat(lexer)
	}
}

// local function Name funcbody	局部函数
func _finishLocalFuncDefStat(lexer *Lexer) *LocalFuncDefStat {
	lexer.NextTokenOfKind(TOKEN_KW_FUNCTION) // local function
	_, name := lexer.NextIdentifier()        // name
	fdExp := parseFuncDefExp(lexer)          // funcbody
	return &LocalFuncDefStat{name, fdExp}
}

// local namelist [‘=’ explist] 局部变量
func _finishLocalVarDeclStat(lexer *Lexer) *LocalVarDeclStat {
	_, name0 := lexer.NextIdentifier()        // local Name
	nameList := _finishNameList(lexer, name0) // { , Name }
	var expList []Exp = nil
	if lexer.LookAhead() == TOKEN_OP_ASSIGN {
		lexer.NextToken()             // ==
		expList = parseExpList(lexer) // explist
	}
	lastLine := lexer.Line()
	return &LocalVarDeclStat{lastLine, nameList, expList}
}

// funcname ::= Name {‘.’ Name} [‘:’ Name]
func _parseFuncName(lexer *Lexer) (exp Exp, hasColon bool) {
	line, name := lexer.NextIdentifier()
	exp = &NameExp{line, name}

	for lexer.LookAhead() == TOKEN_SEP_DOT {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		idx := &StringExp{line, name}
		exp = &TableAccessExp{line, exp, idx}
	}
	if lexer.LookAhead() == TOKEN_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		idx := &StringExp{line, name}
		exp = &TableAccessExp{line, exp, idx}
		hasColon = true
	}

	return
}
