package state

import . "luago/api"

type luaState struct {
	registry *luaTable
	stack    *luaStack
}

// New:创建luaState实例
func New() *luaState {
	registry := newLuaTable(0, 0)
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 0)) // 全局环境
	ls := &luaState{
		registry: registry,
	}
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls)) // 代替了原来用传参或者写死的栈大小
	return ls
}

// 链式调用栈部分

func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}
