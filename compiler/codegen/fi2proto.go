package codegen

import . "luago/binchunk"

func toProto(fi *funcInfo) *Prototype {
	proto := &Prototype{
		NumParams:    byte(fi.numParams),
		MaxStackSize: byte(fi.maxRegs),
		Code:         fi.insts,
		Constants:    getConstants(fi),
		Upvalues:     getUpvalues(fi),
		Protos:       toProtos(fi.subFuncs),
		LineInfo:     []uint32{},
		LocVars:      []LocVar{},
		UpvalueNames: []string{},
	}

	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2 // todo
	}
	if fi.isVararg {
		proto.IsVararg = 1
	}
	return proto
}

func toProtos(fis []*funcInfo) []*Prototype {
	protos := make([]*Prototype, len(fis))
	for i, fi := range fis {
		protos[i] = toProto(fi)
	}
	return protos
}

func getConstants(fi *funcInfo) []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func getUpvalues(fi *funcInfo) []Upvalue {
	upvals := make([]Upvalue, len(fi.upvalues))
	for _, uv := range fi.upvalues {
		// 这里参考indexOfUpval方法中locVarSlot和upvalIndex的定义
		if uv.locVarSlot >= 0 { // upvalue in stack 在栈中
			upvals[uv.index] = Upvalue{1, byte(uv.locVarSlot)}
		} else {
			upvals[uv.index] = Upvalue{0, byte(uv.upvalIndex)}
		}
	}
	return upvals
}
