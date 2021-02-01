package vm

/*
	TODO:还是不太理解，有空研究一下
	将某个字节用二进制写成eeeeexxx
	1. eeeee == 0 则该数表示xxx整数
	2. eeeee != 0 则该字节表示的整数为（1xxx）*2^(eeeee-1)
*/

// 浮点字节流
func Int2fb(x int) int {
	e := 0
	if x < 8 {
		return x
	}
	for x >= (8 << 4) {
		x = (x + 0xf) >> 4 /* x = ceil(x/16) */
		e += 4
	}
	for x >= (8 << 1) {
		x = (x + 1) >> 1
		e++
	}
	return ((e + 1) << 3) | (x - 8)
}

func Fb2int(x int) int {
	if x < 8 {
		return x
	} else {
		return ((x & 7) + 8) << uint((x>>3)-1)
	}
}
