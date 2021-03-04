package ast

type Exp interface {
}

// 简单表达式
type NilExp struct{ Line int }    // nil
type TrueExp struct{ Line int }   // true
type FalseExp struct{ Line int }  // false
type VarargExp struct{ Line int } // ...

// 数字表达式
type IntegerExp struct {
	Line int
	Val  int64
}
type FloatExp struct {
	Line int
	Val  float64
}

// string
type StringExp struct {
	Line int
	Str  string
}

// unop exp 一元表达式
type UnopExp struct {
	Line int // line of operator
	Op   int // operator
	Exp  Exp
}

// exp1 op exp2 二元表达式
type BinopExp struct {
	Line int // line of operator
	Op   int // operator
	Exp1 Exp
	Exp2 Exp
}

// 拼接表达式
type ConcatExp struct {
	Line int // line of last ..
	Exps []Exp
}

// 表构造表达式
// tableconstructor ::= ‘{’ [fieldlist] ‘}’
// fieldlist ::= field {fieldsep field} [fieldsep]
// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
// fieldsep ::= ‘,’ | ‘;’
type TableConstructorExp struct {
	Line     int // line of `{` ?
	LastLine int // line of `}`
	KeyExps  []Exp
	ValExps  []Exp
}

// 函数定义表达式
// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
type FuncDefExp struct {
	Line     int
	LastLine int // line of `end`
	ParList  []string
	IsVararg bool
	Block    *Block
}

// 接下来是前缀表达式
/*
prefixexp ::= Name |
              ‘(’ exp ‘)’ |
              prefixexp ‘[’ exp ‘]’ |
              prefixexp ‘.’ Name |
              prefixexp ‘:’ Name args |
              prefixexp args
*/
// var 表达式
type NameExp struct {
	Line int
	Name string
}

// 圆括号 表达式	改变运算符优先级
type ParensExp struct {
	Exp Exp
}

// 表访问表达式
type TableAccessExp struct {
	LastLine  int // line of `]` ?
	PrefixExp Exp
	KeyExp    Exp
}

// 函数调用表达式
type FuncCallExp struct {
	Line      int // line of `(` ?
	LastLine  int // line of ')'
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
}
