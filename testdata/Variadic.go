package testdata

import "github.com/ymz-ncnk/mok"

type Variadic interface {
	Process(...int)
}

type ProcessFn func(...int)

func NewVariadicMock() VariadicMock {
	return VariadicMock{
		Mock: mok.New("Variadic"),
	}
}

type VariadicMock struct {
	*mok.Mock
}

func (m VariadicMock) RegisterProcess(fn ProcessFn) VariadicMock {
	m.Register("Process", fn)
	return m
}

func (m VariadicMock) Process(args ...int) {
	_, err := m.Call("Process", args)
	if err != nil {
		panic(err)
	}
}
