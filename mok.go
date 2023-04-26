package mok

import (
	"fmt"
	"reflect"
	"sync"
)

// MockName represents a mock name.
type MockName string

// Func represents one method call, should be a function.
type Func interface{}

// New creates new Mock.
func New(name MockName) *Mock {
	return &Mock{name: name}
}

// Mock helps you mock interfaces.
type Mock struct {
	name MockName
	m    sync.Map
}

// Register registers a function as a single method call. You could chain
// Register calls: mock.Register("Handle", ...).Register("Handle", ...).
func (mock *Mock) Register(name MethodName, fn Func) *Mock {
	if !isFunc(fn) {
		panic(ErrNotFunction)
	}
	method, _ := mock.m.LoadOrStore(name, NewMethod())
	method.(*Method).AddMethodCall(fn)
	return mock
}

// RegisterN registers a function as several method calls.
func (mock *Mock) RegisterN(name MethodName, n int, fn Func) *Mock {
	for i := 0; i < n; i++ {
		mock.Register(name, fn)
	}
	return mock
}

// Unregister unregisters all method calls.
func (mock *Mock) Unregister(name MethodName) *Mock {
	mock.m.Delete(name)
	return mock
}

// Call calls a method with the specified params using reflection.
// Params can be of type reflect.Value.
//
// Returns UnknownMethodCallError if no calls was registered,
// UnexpectedMethodCallError - if all registered method calls have already been
// made.
func (mock *Mock) Call(name MethodName, params ...interface{}) (
	[]interface{}, error) {
	method, pst := mock.m.Load(name)
	if !pst {
		return nil, NewUnknownMethodCallError(mock.name, name)
	}
	vals, err := method.(*Method).Call(params)
	if err != nil {
		if err == ErrUnexpectedCall {
			return nil, NewUnexpectedMethodCallError(mock.name, name)
		} else {
			panic(fmt.Sprintf("unepxected '%v' err", err))
		}
	}
	return vals, nil
}

// CheckCalls checks whether all registered method calls of the mock have been
// used. If yes, returns an empty slice.
func (mock *Mock) CheckCalls() []MethodCallsInfo {
	arr := []MethodCallsInfo{}
	mock.m.Range(func(key, value interface{}) bool {
		methodName := key.(MethodName)
		method := value.(*Method)
		info, ok := method.CheckCalls(mock.name, methodName)
		if !ok {
			arr = append(arr, info)
		}
		return true
	})
	return arr
}

func isFunc(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}
