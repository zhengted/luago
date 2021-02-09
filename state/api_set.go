package state

import "luago/api"

func (self *luaState) SetTable(idx int) {
	v := self.stack.pop()
	k := self.stack.pop()
	t := self.stack.get(idx)
	self.setTable(t, k, v, false)
}

func (self *luaState) setTable(t, k, v luaValue, raw bool) {
	if tbl, ok := t.(*luaTable); ok {
		tbl.put(k, v)
		return
	}
	if !raw {
		if mf := getMetafield(t, "__newindex", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				self.setTable(x, k, v, false)
				return
			case *closure:
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.stack.push(v)
				self.Call(3, 0)
				return
			}
		}
	}
	panic("not a table")
}

func (self *luaState) SetField(idx int, k string) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, k, v, false)
}

func (self *luaState) SetI(idx int, k int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, k, v, false)
}

// SetGlobal:栈顶元素弹出，并赋予给一个全局表键值
func (self *luaState) SetGlobal(name string) {
	t := self.registry.get(api.LUA_RIDX_GLOBALS)
	v := self.stack.pop()
	self.setTable(t, name, v, false)
}

// Register:给全局环境注册Go函数值
func (self luaState) Register(name string, f api.GoFunction) {
	self.PushGoFunction(f)
	self.SetGlobal(name)
}

// SetMetatable:将栈顶元素设置为指定索引处的元表，也可以将元表置空
func (self *luaState) SetMetatable(idx int) {
	val := self.stack.get(idx)
	mtVal := self.stack.pop()
	if mtVal == nil {
		setMetatable(val, nil, self)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, self)
	} else {
		panic("table expected")
	}
}

func (self *luaState) RawSet(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, k, v, true)
}

func (self *luaState) RawSetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, true)
}
