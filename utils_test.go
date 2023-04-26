package mok

import (
	"fmt"
	"testing"
)

func TestCheckCalls(t *testing.T) {
	reader := NewReaderMock()
	reader.RegisterRead(func(p []byte) (n int, err error) {
		return 0, nil
	})
	reader.RegisterRead(func(p []byte) (n int, err error) {
		return 1, nil
	})
	reader.Read([]byte{})
	arr := reader.CheckCalls()
	if len(arr) != 1 {
		t.Error("unexpected CheckCalls result")
	}
	err := checkMethodCallsInfo(arr[0], 2, 1)
	if err != nil {
		t.Error(err)
	}
	reader.Read(nil)
	arr = reader.CheckCalls()
	if len(arr) != 0 {
		t.Error("unexpected CheckCalls result")
	}
}

func checkMethodCallsInfo(info MethodCallsInfo, expectedCalls,
	actualCalls int) error {
	if info.MockName != "Reader" {
		return fmt.Errorf("unexpected MockName, want '%v' actual '%v'",
			"Reader",
			info.MockName)
	}
	if info.MethodName != "Read" {
		return fmt.Errorf("unexpected MethodName, want '%v' actual '%v'", "Read",
			info.MethodName)
	}
	if info.ExpectedCalls != expectedCalls {
		return fmt.Errorf("unexpected ExpectedCalls, want '%v' actual '%v'",
			expectedCalls,
			info.ExpectedCalls)
	}
	if info.ActualCalls != actualCalls {
		return fmt.Errorf("unexpected ActualCalls, want '%v' actual '%v'",
			actualCalls,
			info.ActualCalls)
	}
	return nil
}
