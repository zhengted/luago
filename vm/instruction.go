package vm

import (
	"luago/api"
)

const (
	MAXARG_Bx  = 1<<18 - 1      // 2^18 - 1 = 262143
	MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071
)

type Instruction uint32

// Opcode:提取操作码
func (self Instruction) Opcode() int {
	return int(self & 0x3F)
}

// ABC:ABC模式提取操作数
func (self Instruction) ABC() (a, b, c int) {
	a = int(self >> 6 & 0xFF)
	c = int(self >> 14 & 0x1FF)
	b = int(self >> 23 & 0x1FF)
	return
}

// ABx: ABx提取操作数
func (self Instruction) ABx() (a, bx int) {
	a = int(self >> 6 & 0xFF)
	bx = int(self >> 14)
	return
}

// AsBx: AsBx提取操作数
func (self Instruction) AsBx() (a, sbx int) {
	a, bx := self.ABx()
	return a, bx - MAXARG_sBx
}

// 以上两者的区别在于sbx是有符号的，而bx是无符号的
// sbx的取值范围是：-131071~131072
// bx的取值范围是：0~262143

// Ax: Ax提取参数
func (self Instruction) Ax() int {
	return int(self >> 6)
}

// 下面这些比较简单
func (self Instruction) OpName() string {
	return opcodes[self.Opcode()].name
}

func (self Instruction) OpMode() byte {
	return opcodes[self.Opcode()].opMode
}

func (self Instruction) BMode() byte {
	return opcodes[self.Opcode()].argBMode
}

func (self Instruction) CMode() byte {
	return opcodes[self.Opcode()].argCMode
}

func (self Instruction) Execute(vm api.LuaVM) {
	action := opcodes[self.Opcode()].action
	if action != nil {
		action(self, vm)
	} else {
		panic(self.OpName())
	}
}
