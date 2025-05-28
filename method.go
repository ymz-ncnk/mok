package mok

import (
	"fmt"
	"reflect"
	"sync"
)

// MethodName represents a method name.
type MethodName string

// MethodCallsInfo holds an information about method calls.
type MethodCallsInfo struct {
	MockName      MockName
	MethodName    MethodName
	ExpectedCalls int
	ActualCalls   int
}

func (info MethodCallsInfo) String() string {
	return fmt.Sprintf("%v.%v() calls count: want %v, actual %v", info.MockName,
		info.MethodName,
		info.ExpectedCalls,
		info.ActualCalls)
}

// NewMethod creates new Method.
func NewMethod() *Method {
	return &Method{fns: []reflect.Value{}, mu: sync.Mutex{}}
}

// Method represents a struct method.
type Method struct {
	callsCount int
	fns        []reflect.Value
	mu         sync.Mutex
}

// AddMethodCall adds one method call.
func (m *Method) AddMethodCall(fn Func) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fns = append(m.fns, reflect.ValueOf(fn))
}

// Call, using reflection, calls the function that was registered as a method
// call. Params can be of type reflect.Value. Threadsafe.
//
// Returns ErrUnexpectedCall if all registered method calls have already been
// made.
func (m *Method) Call(params []any) ([]any, error) {
	var fn reflect.Value
	m.mu.Lock()
	if len(m.fns) < m.callsCount+1 {
		m.mu.Unlock()
		return nil, ErrUnexpectedCall
	}
	fn = m.fns[m.callsCount]
	m.increaseCallsCount()
	m.mu.Unlock()
	var (
		result []reflect.Value
		vals   = toReflectValues(params)
	)
	if fn.Type().IsVariadic() && vals[len(vals)-1].Kind() == reflect.Slice {
		result = fn.CallSlice(vals)
	} else {
		result = fn.Call(vals)
	}
	return fromReflectValues(result), nil
}

// CheckCalls checks whether all registered method calls have been
// used. If yes, returns ok == true.
func (m *Method) CheckCalls(mockName MockName, methodName MethodName) (
	info MethodCallsInfo, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.fns) != m.callsCount {
		return MethodCallsInfo{mockName, methodName, len(m.fns),
			m.callsCount}, false
	}
	return MethodCallsInfo{}, true
}

func (m *Method) increaseCallsCount() {
	m.callsCount++
}

func toReflectValues(vals []any) []reflect.Value {
	rvals := make([]reflect.Value, len(vals))
	for i := 0; i < len(vals); i++ {
		if rval, ok := vals[i].(reflect.Value); ok {
			rvals[i] = rval
		} else {
			rvals[i] = reflect.ValueOf(vals[i])
		}
	}
	return rvals
}

func fromReflectValues(rvals []reflect.Value) []any {
	vals := make([]any, len(rvals))
	for i := 0; i < len(vals); i++ {
		vals[i] = rvals[i].Interface()
	}
	return vals
}
