package state

// PC:获取当前的PC
func (self *luaState) PC() int {
	return self.stack.pc
}

// AddPC:跳转到指定行指令
func (self *luaState) AddPC(n int) {
	self.stack.pc += n
}

// Fetch: 获取当前指令，并且让指令计数器+1
func (self *luaState) Fetch() uint32 {
	i := self.stack.closure.proto.Code[self.stack.pc]
	self.stack.pc++
	return i
}

// GetConst:从函数原型中获取一个常熟并压入栈中
func (self *luaState) GetConst(idx int) {
	c := self.stack.closure.proto.Constants[idx]
	self.stack.push(c)
}

// GetRK:根据RK值选择将某个常量推入栈顶
//		或者调用PushValue将某个索引处的栈值推入栈顶
func (self *luaState) GetRK(rk int) {
	if rk > 0xFF {
		// 将常量值推入栈顶
		self.GetConst(rk & 0xFF)
	} else {
		// rk相关的在这里已经+1了无需处理
		self.PushValue(rk + 1)
	}
}

// 注意：GetRK的参数是OpArgK类型的，取决于9个比特中的第一个
// 虚拟机指令操作数携带的寄存器索引是从0开始的，lua栈API索引是从1开始的
// 因此寄存器索引当成栈索引使用时要+1

// RegisterCount:返回当前Lua函数所操作的寄存器数量
func (self *luaState) RegisterCount() int {
	return int(self.stack.closure.proto.MaxStackSize)
}

//LoadVararg(n int):传递给当前Lua函数的变长参数推入栈顶
func (self *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(self.stack.varargs)
	}
	self.stack.check(n)
	self.stack.pushN(self.stack.varargs, n)
}

//LoadProto(idx int):把当前Lua函数的子函数原型 实例化为闭包推入栈顶
func (self *luaState) LoadProto(idx int) {
	proto := self.stack.closure.proto.Protos[idx]
	closure := newLuaClosure(proto)
	self.stack.push(closure)
}
