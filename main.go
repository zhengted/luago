package main

import (
	"fmt"
	"io/ioutil"
	. "luago/api"
	"luago/state"
)

func main() {
	// Global Test demo

	data, err := ioutil.ReadFile("luac.out")
	if err != nil {
		panic(err)
	}
	ls := state.New()
	ls.Register("print", print)
	ls.Register("getmetatable", getMetatable)
	ls.Register("setmetatable", setMetatable)
	ls.Load(data, "luac.out", "b")
	ls.Call(0, 0)

}

// 用来注册的go函数
func print(ls LuaState) int {
	nArgs := ls.GetTop()
	for i := 1; i <= nArgs; i++ {
		if ls.IsBoolean(i) {
			fmt.Printf("%t", ls.ToBoolean(i))
		} else if ls.IsString(i) {
			fmt.Print(ls.ToString(i))
		} else {
			fmt.Print(ls.TypeName(ls.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}

func getMetatable(ls LuaState) int {
	if !ls.GetMetatable(1) {
		ls.PushNil()
	}
	return 1
}

func setMetatable(ls LuaState) int {
	ls.SetMetatable(1)
	return 1
}
