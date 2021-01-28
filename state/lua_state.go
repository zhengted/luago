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
