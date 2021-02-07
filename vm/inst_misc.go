package vm

import . "luago/api"

// move:R(A) = R(B)
func move(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1 // 寄存器索引+1才是相应的栈索引
	b += 1
	vm.Copy(b, a)
}

// jmp:pc跳转
func jmp(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	vm.AddPC(sBx)
	if a != 0 {
		// 处于开启状态的Upvalue引用了还在寄存器里的Lua值，
		// 将这些值从寄存器里复制出来，然后更新Upvalue
		vm.CloseUpvalues(a)
	}
}
