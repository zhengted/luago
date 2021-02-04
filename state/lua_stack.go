package state

// luaStack:lua栈目前定义 栈+top
type luaStack struct {
	slots   []luaValue
	top     int
	prev    *luaStack  // prev:与函数执行没有关系，让调用帧变成链表结点
	closure *closure   // closure:闭包，可以理解成函数原型
	varargs []luaValue // varargs:变长参数，
	pc      int        // pc:指令计数器
}

// newLuaStack:工厂创建lua栈
func newLuaStack(size int) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
	}
}

// check:检查栈的空闲空间是否还可以容纳至少n个值，如果不满足这个条件，则调用append扩容
func (self *luaStack) check(n int) {
	free := len(self.slots) - self.top
	for i := free; i < n; i++ {
		self.slots = append(self.slots, nil)
	}
}

// push: 压入栈顶，失败则panic
func (self *luaStack) push(val luaValue) {
	if self.top == len(self.slots) {
		panic("stack overflow")
	}
	self.slots[self.top] = val
	self.top++
}

// pushN:将多个值（luaValue）压入栈顶，n和len(luaValue)多退少补
func (self *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}
	for i := 0; i < n; i++ {
		if i < nVals {
			self.push(vals[i])
		} else {
			self.push(nil)
		}
	}
}

// pop: 弹出，返回栈顶元素
func (self *luaStack) pop() luaValue {
	if self.top < 1 {
		panic("stack underflow")
	}
	self.top--
	val := self.slots[self.top]
	self.slots[self.top] = nil // 如何正确处理删除切片元素，这里如果使用切片移动指针的方式，会造成内存泄漏，因为切片为接口类型切片
	return val
}

// popN: 弹出栈顶指定数量的值
func (self *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := 0; i < n; i++ {
		vals[i] = self.pop()
	}
	return vals
}

// absIndex: 把索引转换为绝对索引
// TODO:需考虑索引是否有效
func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 {
		return idx
	}
	return idx + self.top + 1
}

// isValid: 判断索引是否有效
func (self *luaStack) isValid(idx int) bool {
	absIdx := self.absIndex(idx)
	return absIdx > 0 && absIdx <= self.top
}

// get: 根据索引从栈中取值
func (self *luaStack) get(idx int) luaValue {
	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		return self.slots[absIdx-1]
	}
	return nil
}

// set: 根据索引往栈中写值
func (self *luaStack) set(idx int, val luaValue) {
	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		self.slots[absIdx-1] = val
		return
	}
	panic("invalid index")
}

// reverse:将slot从索引idx1到idx2翻转
func (self *luaStack) reverse(from, to int) {
	slots := self.slots
	// reverse模板
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}
