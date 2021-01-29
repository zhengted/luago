package state

type luaTable struct {
	arr  []luaValue            // lua表数组部分
	_map map[luaValue]luaValue // lua表哈希部分
}

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
