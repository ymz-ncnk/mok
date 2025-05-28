package mok

import (
	"fmt"
	"reflect"
	"sync"
)

// MockName represents a mock name.
type MockName string

// Func represents one method call, should be a function.
type Func any

// New creates new Mock.
func New(name MockName) *Mock {
	return &Mock{name: name}
}

// Mock helps you mock interfaces.
type Mock struct {
	name    MockName
	syncMap sync.Map
}

// Register registers a function as a single method call. You could chain
// Register calls: mock.Register("Handle", ...).Register("Handle", ...).
func (m *Mock) Register(name MethodName, fn Func) *Mock {
	if !isFunc(fn) {
		panic(ErrNotFunction)
	}
	method, _ := m.syncMap.LoadOrStore(name, NewMethod())
	method.(*Method).AddMethodCall(fn)
	return m
}

// RegisterN registers a function as several method calls.
func (m *Mock) RegisterN(name MethodName, n int, fn Func) *Mock {
	for i := 0; i < n; i++ {
		m.Register(name, fn)
	}
	return m
}

// Unregister unregisters all method malls.
func (m *Mock) Unregister(name MethodName) *Mock {
	m.syncMap.Delete(name)
	return m
}

// Call calls a method with the specified params using reflection.
// Params can be of type reflect.Value.
//
// Returns UnknownMethodCallError if no calls was registered,
// UnexpectedMethodCallError - if all registered method calls have already been
// made.
func (m *Mock) Call(name MethodName, params ...any) ([]any, error) {
	method, pst := m.syncMap.Load(name)
	if !pst {
		return nil, NewUnknownMethodCallError(m.name, name)
	}
	vals, err := method.(*Method).Call(params)
	if err != nil {
		if err == ErrUnexpectedCall {
			return nil, NewUnexpectedMethodCallError(m.name, name)
		} else {
			panic(fmt.Sprintf("unepxected '%v' err", err))
		}
	}
	return vals, nil
}

// CheckCalls checks whether all registered method calls of the mock have been
// used. If yes, returns an empty slice.
func (m *Mock) CheckCalls() []MethodCallsInfo {
	arr := []MethodCallsInfo{}
	m.syncMap.Range(func(key, value interface{}) bool {
		methodName := key.(MethodName)
		method := value.(*Method)
		info, ok := method.CheckCalls(m.name, methodName)
		if !ok {
			arr = append(arr, info)
		}
		return true
	})
	return arr
}

func isFunc(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}
