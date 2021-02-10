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

// GetTable: 取栈顶的表索引值的对应value，置入栈顶，参数为栈索引指向的表，原来的表会出栈
func (self *luaState) GetTable(idx int) LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, false)
}

func (self *luaState) getTable(t, k luaValue, raw bool) LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			self.stack.push(v)
			return typeOf(v)
		}
	}
	if !raw {
		if mf := getMetafield(t, "__index", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return self.getTable(x, k, false)
			case *closure:
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.Call(2, 1)
				v := self.stack.get(-1)
				return typeOf(v)
			}
		}
	}
	panic("index error")
}

func (self *luaState) GetField(idx int, k string) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, k, false)
	/*
		以下方法也可：
			self.PushString(k)
			return self.GetTable(idx)
	*/

}

func (self *luaState) GetI(idx int, i int64) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, false)
}

// GetGlobal:将全局表中字段名为name的变量push入栈
func (self *luaState) GetGlobal(name string) LuaType {
	t := self.registry.get(LUA_RIDX_GLOBALS)
	return self.getTable(t, name, false)
}

// GetMetatable:获取指定索引处的元表
func (self *luaState) GetMetatable(idx int) bool {
	val := self.stack.get(idx)
	if mt := getMetatable(val, self); mt != nil {
		self.stack.push(mt)
		return true
	} else {
		return false
	}
}

func (self *luaState) RawGet(idx int) LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, true)
}

func (self *luaState) RawGetI(idx int, i int64) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, true)
}
