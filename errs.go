package mok

import (
	"errors"
	"fmt"
)

// ErrNotFunction happens during the registration of an object that is not a
// function.
var ErrNotFunction = errors.New("not a function")

// ErrUnexpectedCall happens during an unexpected method call.
var ErrUnexpectedCall = errors.New("unexpected call")

// NewUnexpectedMethodCallError creates new UnexpectedMethodCallError.
func NewUnexpectedMethodCallError(mockName MockName,
	methodName MethodName) *UnexpectedMethodCallError {
	return &UnexpectedMethodCallError{mockName, methodName}
}

// UnexpectedMethodCallError happens during an unexpected method call.
type UnexpectedMethodCallError struct {
	mockName   MockName
	methodName MethodName
}

func (err *UnexpectedMethodCallError) MockName() MockName {
	return err.mockName
}

func (err *UnexpectedMethodCallError) MethodName() MethodName {
	return err.methodName
}

func (err *UnexpectedMethodCallError) Error() string {
	return fmt.Sprintf("unexpected %s.%s() method call", err.mockName,
		err.methodName)
}

// NewUnknownMethodCallError creates new UnknownMethodCallError.
func NewUnknownMethodCallError(mockName MockName,
	methodName MethodName) *UnknownMethodCallError {
	return &UnknownMethodCallError{mockName, methodName}
}

// UnknownMethodCallError happens during an unregistered method call.
type UnknownMethodCallError struct {
	mockName   MockName
	methodName MethodName
}

func (err *UnknownMethodCallError) MockName() MockName {
	return err.mockName
}

func (err *UnknownMethodCallError) MethodName() MethodName {
	return err.methodName
}

func (err *UnknownMethodCallError) Error() string {
	return fmt.Sprintf("unknown %s.%s() method call", err.mockName,
		err.methodName)
}
