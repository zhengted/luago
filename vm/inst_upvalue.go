package vm

import (
	. "luago/api"
	//. "luago/state"
)

// getTabUp: R(A) := UpValue[B][RK(C)]
//	先用RK将指定键推入栈顶，调用GetTable方法获得Upvalue中的表，最后调用replace
//			适用于upvalue中存的是table类型
func getTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.GetRK(c)
	vm.GetTable(LuaUpvalueIndex(b))
	vm.Replace(a)
}

// getUpval:R(A) := UpValue[B] 获取指定索引的uv（b）复制到指定寄存器（a）
//		注意：Upval索引在虚拟机的操作数里从0开始，转换成lua栈伪索引时是从1开始的
func getUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Copy(LuaUpvalueIndex(b), a)
}

// setUpval:Upvalue[B] := R(A)将指定寄存器的值a赋值到指定索引的uv
func setUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Copy(a, LuaUpvalueIndex(b))
}

// setTabUp:UpValue[A][RK(B)] := RK(C)
func setTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(LuaUpvalueIndex(a))
}
