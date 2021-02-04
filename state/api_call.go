package state

import (
	"fmt"
	"luago/binchunk"
)
import "luago/vm"

/*
	Load: 加载chunk，可以是lua也可以是编译后的二进制chunk，根据mode来决定
		返回值为状态码：0表示成功
*/
func (self *luaState) Load(chunk []byte, chunkName, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	self.stack.push(c)
	return 0
}

// callLuaClosure:具体逻辑，
func (self *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	// 1. 初始化信息，确定寄存器的数量，定义函数时声明的固定参数数量、
	//	以及是否是vararg函数（会是当扩大）
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	// 2. 创建一个新的调用帧，把闭包（函数原型）和调用帧联系起来
	newStack := newLuaStack(nRegs + 20)
	newStack.closure = c

	// 3. 调用popN把函数和参数值一次性从栈顶弹出，
	//	然后调用新帧的pushN方法按照固定参数数量传入参数
	funcAndArgs := self.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	// 4. 将新帧push进调用栈栈顶，让他成为当前帧，最后调用runLuaClosure
	self.pushLuaStack(newStack)
	self.runLuaClosure()
	self.popLuaStack()

	// 5. 将结果压入旧的调用栈中
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		self.stack.push(len(results)) // 草（一种植物），结果在被压入栈时会先压一个结果长度入栈
		self.stack.pushN(results, nResults)
	}
}

// runLuaClosure:调用栈顶函数
func (self *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(self.Fetch())
		inst.Execute(self)
		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}

// Call: 函数调用
// 	参数说明：nArgs 参数在寄存器中的索引  nResult：结果值的初始索引（因为会有多个返回值）
//	也可以理解成被调函数的在寄存器中的索引
func (self *luaState) Call(nArgs, nResults int) {
	val := self.stack.get(-(nArgs + 1))
	if c, ok := val.(*closure); ok {
		fmt.Printf("call %s<%d,%d>\n", c.proto.Source, c.proto.LineDefined,
			c.proto.LastLineDefined)
		self.callLuaClosure(nArgs, nResults, c)
	} else {
		panic("not function")
	}
}
