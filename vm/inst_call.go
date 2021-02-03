package vm

import . "luago/api"

// closure: ABx，R(A) := closure(KPROTO[Bx]) 将Bx所指向的原型表中的内容压入寄存器中
//		把当前Lua函数的子函数原型实例化为闭包
func closure(i Instruction, vm LuaVM) {
	a, bx := i.ABx()
	a += 1
	vm.LoadProto(bx)
	vm.Replace(a)
}

// call:iABC,R(A), ... , R(A+C-2) := R(A)(R(A+1),...,R(A+B-1))
//  具体操作看2625的图片
func call(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResult(a, c, vm)
}

// _return:
func _return(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	if b == 1 {
		// 无返回值
	} else if b > 1 {
		// 有b-1个返回值
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		_fixStack(a, vm)
	}
}

// vararg:R(A),R(A+1),...,R(A+B-2) = vararg
func vararg(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	if b != 1 {
		vm.LoadVararg(b - 1)
		_popResult(a, b, vm)
	}
}

// TailCall:尾调用 return f(args)
// 		return R(A)(R(A+1),...,R(A+B-1))
func tailCall(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	c := 0
	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResult(a, c, vm)
}

// self:R(A+1) := R(B); R(A) := R(B)[RK(c)]
func self(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.Copy(b, a+1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

// TODO:这块的内容需要画图理解，暂时还没搞太懂，花时间再看一遍书 21.2.3

// _pushFuncAndArgs:
func _pushFuncAndArgs(a, b int, vm LuaVM) int {
	if b-1 >= 0 {
		vm.CheckStack(b)
		// b>1 按序入栈即可
		for i := a; i < a+b; i++ {
			vm.PushValue(i)
		}
		return b - 1
	} else {
		_fixStack(a, vm)
		return vm.GetTop() - vm.RegisterCount() - 1
	}
}

// _fixStack
func _fixStack(a int, vm LuaVM) {
	x := int(vm.ToInteger(-1))
	vm.Pop(1)

	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

// _popResult
func _popResult(a, c int, vm LuaVM) {
	if c == 1 {
		// no result
		// 无返回值
	} else if c > 1 {
		// 按序替换栈指定寄存器内容即可
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else {
		// c == 0 让返回值们留在栈顶，再往其中推入一个目标寄存器的索引入栈
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}
