package state

import (
	. "luago/api"
)

// CreateTable: 建表，带长度的
func (self *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	self.stack.push(t)
}

// NewTable: 新建表，数组和哈希长度都为0
func (self *luaState) NewTable() {
	t := newLuaTable(0, 0)
	self.stack.push(t)
}

// GetTable: 取栈顶的表索引值的对应value，置入栈顶，参数为栈索引指向的表
func (self *luaState) GetTable(idx int) LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k)
}

func (self *luaState) getTable(t, k luaValue) LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		self.stack.push(v)
		return typeOf(v)
	}
	panic("Not a table")
}

func (self *luaState) GetField(idx int, k string) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, k)
	/*
		以下方法也可：
			self.PushString(k)
			return self.GetTable(idx)
	*/

}

func (self *luaState) GetI(idx int, i int64) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i)
}
