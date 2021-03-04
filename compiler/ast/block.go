package ast

type Block struct {
	LastLine int
	Stats    []Stat
	RetExps  []Exp
}
