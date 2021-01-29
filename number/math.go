package number

import (
	"math"
)

// FloatToInteger:浮点数转整数
// todo:思考什么情况下会出现错误
func FloatToInteger(f float64) (int64, bool) {
	i := int64(f)
	return i, float64(i) == f
}

// IFloorDiv:整数类型除法
func IFloorDiv(a, b int64) int64 {
	if a > 0 && b > 0 || a < 0 && b < 0 || a%b == 0 {
		return a / b
	} else {
		return a/b - 1
	}
}

// FFloorDiv:浮点数类型除法
func FFloorDiv(a, b float64) float64 {
	return math.Floor(a / b)
}

// IMod: 整数类型取模
// a % b == a - ((a // b) * b)
func IMod(a, b int64) int64 {
	return a - IFloorDiv(a, b)*b
}

// FMod: 浮点数类型取模
func FMod(a, b float64) float64 {
	// 第二版增加了一个判断是否为无穷值
	if a > 0 && math.IsInf(b, 1) || a < 0 && math.IsInf(b, -1) {
		return a
	}
	if a > 0 && math.IsInf(b, -1) || a < 0 && math.IsInf(b, 1) {
		return b
	}
	return a - math.Floor(a/b)*b
}

// ShiftRight: 右移函数
func ShiftRight(a, n int64) int64 {
	if n >= 0 {
		return int64(uint64(a) >> uint64(n))
	} else {
		return ShiftLeft(a, -n)
	}
}

// ShiftLeft: 左移函数
func ShiftLeft(a, n int64) int64 {
	if n >= 0 {
		return int64(uint64(a) << uint64(n))
	} else {
		return ShiftRight(a, -n)
	}
}
