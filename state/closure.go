package state

import (
	. "luago/api"
	"luago/binchunk"
)

type closure struct {
	// 约定proto为nil则为go闭包  gofunc为nil则为Lua闭包
	proto  *binchunk.Prototype
	goFunc GoFunction
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	return &closure{
		proto: proto,
	}
}

func newGoClosure(f GoFunction) *closure {
	return &closure{
		goFunc: f,
	}
}
