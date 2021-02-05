package state

import (
	"fmt"
	"luago/api"
)

func (self *luaState) PushNil() {
	self.stack.push(nil)
}

func (self *luaState) PushBoolean(b bool) {
	self.stack.push(b)
}

func (self *luaState) PushInteger(n int64) {
	self.stack.push(n)
}

func (self *luaState) PushNumber(n float64) {
	self.stack.push(n)
}

func (self *luaState) PushString(s string) {
	self.stack.push(s)
}

// 测试用
func (self *luaState) PrintStack() {
	fmt.Printf("PrintStack %v\n", self.stack.slots)
}

func (self *luaState) PushGoFunction(f api.GoFunction) {
	self.stack.push(newGoClosure(f))
}

// PushGlobalTable:将全局表push进栈
func (self *luaState) PushGlobalTable() {
	// 取出注册表，取出注册表中的全局表
	global := self.registry.get(api.LUA_RIDX_GLOBALS)
	self.stack.push(global)

	// 可用下面这句替换
	// self.GetI(api.LUA_REGISTRYINDEX, api.LUA_RIDX_GLOBALS)
}
