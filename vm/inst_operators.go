package vm

import . "luago/api"

// 运算符运算的具体实现
func add(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPADD) }  // +
func sub(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSUB) }  // -
func mul(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPMUL) }  // *
func mod(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPMOD) }  // %
func pow(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPPOW) }  // ^
func div(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPDIV) }  // /
func idiv(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPIDIV) } // //
func band(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPBAND) } // &
func bor(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPBOR) }  // |
func bxor(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPBXOR) } // ~
func shl(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSHL) }  // <<
func shr(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSHR) }  // >>
func unm(i Instruction, vm LuaVM)  { _unaryArith(i, vm, LUA_OPUNM) }   // -
func bnot(i Instruction, vm LuaVM) { _unaryArith(i, vm, LUA_OPBNOT) }  // ~

/*************************** 运算符相关 **************************/
// _binaryArith: R(A) := RK(B) op RK(C)
func _binaryArith(i Instruction, vm LuaVM, op ArithOp) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b) // 将指定（常量或寄存器索引的值）推入栈顶
	vm.GetRK(c)
	vm.Arith(op) // 二元运算并将结果赋给栈顶
	vm.Replace(a)
}

// _unaryArith: R(A) := op R(B)
func _unaryArith(i Instruction, vm LuaVM, op ArithOp) {
	a, b, _ := i.ABC() // 依然是ABC模式 -_-|||
	a += 1
	b += 1

	vm.PushValue(b)
	vm.Arith(op)
	vm.Replace(a)
}

func eq(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPEQ) }
func lt(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPLT) }
func le(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPLE) }

/*************************** 比较相关 **************************/
// _compare: if((RK(B) op RK(c)) ~= A) then pc++
// 这条指令也是没有改变寄存器状态的指令
func _compare(i Instruction, vm LuaVM, op CompareOp) {
	a, b, c := i.ABC()
	// A == 0 false
	// A != 0 true
	vm.GetRK(b)
	vm.GetRK(c)
	if vm.Compare(-2, -1, op) != (a != 0) {
		vm.AddPC(1)
	}
	vm.Pop(2)
}

/*************************** 逻辑运算相关 **************************/

// not:R(A) := not R(B)	只针对boolean值
func not(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a)
}

// test:if not (bool(R(A)) == C) then pc++
func test(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1
	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}

// testSet: if (bool(R(B)) == c) then R(A) := R(B) else PC++
func testSet(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	//res := vm.ToBoolean(b)
	//if res == (c != 0) {		// 错误写法：这里改变了B寄存器的值
	//	vm.PushBoolean(res)
	//	vm.Replace(a)
	//} else {
	//	vm.AddPC(1)
	//}
	// 正解
	if vm.ToBoolean(b) == (c != 0) {
		vm.Copy(b, a)
	} else {
		vm.AddPC(1)
	}
}

/*************************** 长度计算相关 **************************/
// _len: R(A) := length of R(B) 计算长度
func length(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Len(b)
	vm.Replace(a)
}

/*************************** 字符串连接相关 **************************/
// concat: R(A) := R(B).. ... ..R(C) 字符串拼接
func concat(i Instruction, vm LuaVM) {
	/*
		注意 这条指令没有出栈操作
			但是需要先将会被连接的lua值入栈 c-b+1次
			然后再还原(Concat带了一次出栈操作)
	*/
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1

	n := c - b + 1
	vm.CheckStack(n) // 入栈数量不确定需要做检查
	// 21.2.10 入栈的时候遗漏了将R(C)入栈
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n)
	vm.Replace(a)
}
