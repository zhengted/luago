package vm

import . "luago/api"

// loadNil:R(A),R(A+1),...R(A+B) := nil
func loadNil(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	vm.PushNil() // 先入栈一个nil值
	for i := a; i <= a+b; i++ {
		vm.Copy(-1, i) // 将该nil值朝指定位置赋值
	}
	vm.Pop(1) // 弹出循环之前的nil值，保证栈顶指针和执行前一直

	// 思考：如果采用循环push会影响栈顶指针 不可取

}

// loadBool:R(A) := bool(B) if(C) pc++
//			寄存器索引  布尔值		是否跳转
func loadBool(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.PushBoolean(b != 0)
	vm.Replace(a) // 和loadNil的区别 replace内置了出栈操作 无需增加出栈的代码
	if c != 0 {
		vm.AddPC(1)
	}
}

// LoadK:R(A) := Kst(bX) 寄存器索引由A决定，常量表索引由Bx决定
func loadK(i Instruction, vm LuaVM) {
	a, bx := i.ABx()
	a += 1
	vm.GetConst(bx)
	vm.Replace(a)
}

// LoadKX: 为了防止bx超过限度新增一层索引，
//		bx占用18个比特，能表示的最大无符号整数是262143
func loadKx(i Instruction, vm LuaVM) {
	a, _ := i.ABx()
	a += 1

	// 这个指令表能索引到常量表的范围更大
	ax := Instruction(vm.Fetch()).Ax() // iAx模式
	vm.GetConst(ax)
	vm.Replace(a)
}
