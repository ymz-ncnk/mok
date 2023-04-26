package mok

// CheckCalls checks whether all registered method calls were used for each
// mock. If yes, it returns an empty infomap. Otherwise, returns an infomap
// where the key is the index of the mocks parameter, and the value is
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
