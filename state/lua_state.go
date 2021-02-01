package state

type luaState struct {
	stack *luaStack
}

// New:创建luaState实例
func New() *luaState {
	return &luaState{
		stack: newLuaStack(20),
	}
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
