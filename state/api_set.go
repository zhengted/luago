package state

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
