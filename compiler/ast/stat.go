package ast

type Stat interface {
}

// 简单语句
type EmptyStat struct{}              // `;`	无任何语义 分割作用
type BreakStat struct{ Line int }    // break 跳转指令，记录行号
type LabelStat struct{ Name string } // `::` Name `::`
type GotoStat struct{ Name string }  // goto Name
type DoStat struct{ Block *Block }   // do Block end
type FuncCallStat = FuncCallExp      // function call

// 循环语句
// while exp do block end
// repeat block until exp
type WhileStat struct {
	Exp   Exp
	Block *Block
}

type RepeatStat struct {
	Block *Block
	Exp   Exp
}

// 条件语句
// if exp then block ( elseif exp then block ) end
type IfStat struct {
	Exps   []Exp
	Blocks []*Block
}

// 数值For循环语句
// for Name '=' exp ',' [',' exp] do block end
type ForNumStat struct {
	LineOfFor int
	LineOfDo  int
	VarName   string
	InitExp   Exp
	LimitExp  Exp
	StepExp   Exp
	Block     *Block
}

// 通用For循环语句
// for namelist in explist do block end
// namelist :: = Name {',' Name}
// explist :: = exp {',' exp}
type ForInStat struct {
	LineOfDo int
	NameList []string
	ExpList  []Exp
	Block    *Block
}

// 局部变量声明语句
// local namelist ['=' explist]
// namelist :: = Name{',' Name}
// explist :: = exp{',' exp}
type LocalVarDeclStat struct {
	LastLine int
	NameList []string
	ExpList  []Exp
}

// 赋值语句
// varlist '=' explist
// varlist ::= var {',' var}
// var ::= Name | prefixexp '[' exp ']' | prefixexp '.' Name
// explist ::= exp{',' exp}
type AssignStat struct {
	LastLine int
	VarList  []Exp
	ExpList  []Exp
}

// 局部函数定义语句
// local function Name funcbody
type LocalFuncDefStat struct {
	Name string
	Exp  *FuncDefExp
}

// 非局部函数定义语句
// function funcname funcbody
// funcname ::= Name{'.' Name} [':' Name]
// funcbody ::= '(' [parlist] ')' block end
