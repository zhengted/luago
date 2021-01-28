package state

//GetTop() int
//AbsIndex(idx int) int
//CheckStack(n int) bool
//Pop(n int)
//Copy(fromIdx, toIdx int)
//PushValue(idx int)
//Replace(idx int)
//Insert(idx int)
//Remove(idx int)
//Rotate(idx, n int)
//SetTop(idx int)

// GetTop:获取栈顶元素
func (self *luaState) GetTop() int {
	return self.stack.top
}

// AbsIndex:获取绝对索引
func (self *luaState) AbsIndex(idx int) int {
	return self.stack.absIndex(idx)
}

// CheckStack:检查栈是否允许扩容
func (self *luaState) CheckStack(n int) bool {
	self.stack.check(n)
	return true // 暂时认为扩容总是成功
}

// Pop:弹出栈顶的n个元素
func (self *luaState) Pop(n int) {
	self.SetTop(-n - 1)
}

// Copy:复制指定的索引到另一个指定索引
func (self *luaState) Copy(fromIdx, toIdx int) {
	val := self.stack.get(fromIdx)
	self.stack.set(toIdx, val)
}

// PushValue: 将指定索引的值压入栈顶
func (self *luaState) PushValue(idx int) {
	val := self.stack.get(idx)
	self.stack.push(val)
}

// Replace: 将栈顶值弹出并赋值到指定索引
func (self *luaState) Replace(idx int) {
	val := self.stack.pop()
	self.stack.set(idx, val)
}

// Insert: 将栈顶值弹出并插入到指定索引
func (self *luaState) Insert(idx int) {
	self.Rotate(idx, 1) // 将idx以上的部分向上旋转一个单位
}

// Remove: 删除指定索引的值，并将该索引上面的元素下移一个单位
func (self *luaState) Remove(idx int) {
	self.Rotate(idx, -1)
	self.Pop(1)
}

// Rotate: 从指定索引开始旋转栈n个单位，只会影响指定索引上方的元素
func (self *luaState) Rotate(idx, n int) {
	t := self.stack.top - 1
	p := self.stack.absIndex(idx) - 1
	var m int
	if n >= 0 {
		// 向上旋转
		m = t - n
	} else {
		// 向下旋转
		m = p - n - 1
	}
	self.stack.reverse(p, m)
	self.stack.reverse(m+1, t)
	self.stack.reverse(p, t)
}

// SetTop: 将栈顶索引设置为指定值
func (self *luaState) SetTop(idx int) {
	newTop := self.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow")
	}

	n := self.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			self.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			self.stack.push(nil)
		}
	}
}
