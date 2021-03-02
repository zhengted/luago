package vm

import (
	. "luago/api"
)

// closure: ABx，R(A) := closure(KPROTO[Bx]) 把当前Lua函数的子原型实例化为闭包。
//	从当前闭包的子函数原型表中取出原型（用Bx去索引），实例化为闭包，
//	并推入LuaState栈顶，再从栈顶弹出赋值到指定寄存器（A）上
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

// _return:A表示返回值第一项的栈索引，通过B能计算数量
// 把存放在连续多个寄存器里的值返回给主调函数。A决定第一个寄存器索引，寄存器数量由B决定
func _return(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	if b == 1 {
		// 无返回值
	} else if b > 1 {
		// 有b-1个返回值E
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		// 经过函数调用的已经在栈顶了，将另一部分压入栈然后旋转
		_fixStack(a, vm)
	}
}

// vararg:R(A),R(A+1),...,R(A+B-2) = vararg 对应Lua脚本里的...
// 		把传递给当前函数的变长参数加载到连续多个寄存器中。第一个寄存器由A指定，寄存器数量由B指定
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
// 把对象和方法拷贝到相邻的两个目标寄存器中，
//	对象本身在寄存器中，索引由操作数B决定。方法名在常量表中，索引由操作数决定
func self(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.Copy(b, a+1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

// _pushFuncAndArgs:确定可变参数的数量
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

// _fixStack：调整可变参数的位置
func _fixStack(a int, vm LuaVM) {
	// 先将栈顶取出来，此时栈顶表示的是，变长参数所在的寄存器索引
	x := int(vm.ToInteger(-1))
	vm.Pop(1)

	// 常规的将函数和参数压入栈
	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}

	// 将变长参数挪到栈顶
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

// _popResult：
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
		// c == 0 让返回值们留在栈中，再往其中推入一个目标寄存器的索引入栈
		// 如果被调用到就先将目标寄存器的索引取出 再旋转栈
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}

// tForCal: 通用for循环的实现
//  R(A+3),...,R(A+2+C) := R(A)(R(A+1),R(A+2))
func tForCall(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1

	_pushFuncAndArgs(a, 3, vm)
	vm.Call(2, c)
	_popResult(a+3, c+1, vm)
}

// tForLoop: 通用for循环的实现
// if R(A+1) ~= NIL then {
//		R(A) = R(A+1); PC += sBx
// }
func tForLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}
