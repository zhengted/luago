package state

import (
	"luago/number"
	"math"
)

type luaTable struct {
	arr       []luaValue            // lua表数组部分
	_map      map[luaValue]luaValue // lua表哈希部分
	metatable *luaTable             // 元表支持
}

// newLuaTable:新建Lua表
func newLuaTable(nArr, nRec int) *luaTable {
	t := &luaTable{}
	if nArr > 0 {
		t.arr = make([]luaValue, nArr)
	}
	if nRec > 0 {
		t._map = make(map[luaValue]luaValue, nRec)
	}
	return t
}

// get: 根据键值从表里查找值
func (self *luaTable) get(key luaValue) luaValue {
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok {
		if idx >= 1 && idx <= int64(len(self.arr)) {
			return self.arr[idx-1]
		}
	}
	return self._map[key]
}

func _floatToInteger(key luaValue) luaValue {
	if f, ok := key.(float64); ok {
		if i, ok := number.FloatToInteger(f); ok {
			return i
		}
	}
	return key
}

func (self *luaTable) put(key, val luaValue) {
	if nil == key {
		panic("lua table:wrong key")
	}
	if f, ok := key.(float64); ok && math.IsNaN(f) {
		panic("table index is nil")
	}
	if idx, ok := key.(int64); ok && idx >= 1 {
		arrLen := int64(len(self.arr))
		if idx <= arrLen {
			self.arr[idx-1] = val
			if arrLen == idx && nil == val {
				// 如果在末尾，将函数尾部的hole全部删除
				self._shrinkArray()
			}
			return
		}
		if arrLen+1 == idx {
			delete(self._map, key)
			if val != nil {
				// 在末尾后一位则扩展数组部分
				self.arr = append(self.arr, val)
				self._expandArray()
				/*
					这里举个例子：
						如果数组长度一开始是2 并且定义了key为1和2的值
						哈希部分存了数值4，5
						如果定义了key为3的值，则将哈希部分的4，5塞入数组部分
				*/

			}
			return
		}
	}
	// 哈希部分
	if val != nil {
		if self._map == nil {
			self._map = make(map[luaValue]luaValue, 8)
		}
		self._map[key] = val
	} else {
		delete(self._map, key)
	}
}

// _shrinkArray: 删除数组中多余的hole（值为nil的key）
func (self *luaTable) _shrinkArray() {
	for i := len(self.arr) - 1; i >= 0; i-- {
		if self.arr[i] == nil {
			self.arr = self.arr[0:i]
		}
	}
}

// _expandArray: 数组动态扩展
func (self *luaTable) _expandArray() {
	for idx := int64(len(self.arr)) + 1; true; idx++ {
		if val, found := self._map[idx]; found {
			delete(self._map, idx)
			self.arr = append(self.arr, val)
		} else {
			break
		}
	}
}

// len:这个方法和项目中的table_count 不是一个方法，理解成
//		table.getn方法，只计算数组部分长度
func (self *luaTable) len() int {
	return len(self.arr)
}

// hasMetafield:是否有元方法
func (self *luaTable) hasMetafield(fieldName string) bool {
	return self.metatable != nil && self.metatable.get(fieldName) != nil
}
