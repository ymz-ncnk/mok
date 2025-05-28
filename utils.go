package mok

import "reflect"

// EmptyInfomap is a predefined empty infomap.
var EmptyInfomap = map[int][]MethodCallsInfo{}

// CheckCalls checks whether all registered method calls were used for each
// mock. If yes, it returns an empty infomap. Otherwise, returns an infomap
// where the key indicates the element of the mocks slice, and the value is
// a set of MethodCallsInfo.
func CheckCalls(mocks []*Mock) (infomap map[int][]MethodCallsInfo) {
	infomap = make(map[int][]MethodCallsInfo)
	for i := 0; i < len(mocks); i++ {
		info := mocks[i].CheckCalls()
		if len(info) > 0 {
			infomap[i] = info
		}
	}
	return
}

// SafeVal creates a safe value that can be used in mok.Call() call.
//
// We have to use this function because calling the mok.Call() method with a
// nil parameter can cause a panic like:
// panic: reflect: Call using zero Value argument...
func SafeVal[T any](i any) (v reflect.Value) {
	v = reflect.ValueOf(i)
	if i == nil {
		v = reflect.Zero(reflect.TypeOf((*T)(nil)).Elem())
	}
	return
}
