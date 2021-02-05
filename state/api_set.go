package state

import "luago/api"

func (self *luaState) SetTable(idx int) {
	v := self.stack.pop()
	k := self.stack.pop()
	t := self.stack.get(idx)
	self.setTable(t, k, v)
}

func (self *luaState) setTable(t, k, v luaValue) {
	if tbl, ok := t.(*luaTable); ok {
		tbl.put(k, v)
		return
	}
	panic("not a table")
}

func (self *luaState) SetField(idx int, k string) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, k, v)
}

func (self *luaState) SetI(idx int, k int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, k, v)
}

// SetGlobal:栈顶元素弹出，并赋予给一个全局表键值
func (self *luaState) SetGlobal(name string) {
	t := self.registry.get(api.LUA_RIDX_GLOBALS)
	v := self.stack.pop()
	self.setTable(t, name, v)
}

// Register:给全局环境注册Go函数值
func (self luaState) Register(name string, f api.GoFunction) {
	self.PushGoFunction(f)
	self.SetGlobal(name)
}
