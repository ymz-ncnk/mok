package mok_test

import (
	"fmt"
	"testing"

	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
	"github.com/ymz-ncnk/mok/testdata"
)

func TestCheckCalls(t *testing.T) {
	reader := testdata.NewReaderMock().RegisterRead(
		func(p []byte) (n int, err error) {
			return 0, nil
		},
	).RegisterRead(
		func(p []byte) (n int, err error) {
			return 1, nil
		},
	)
	reader.Read([]byte{})
	arr := reader.Mock.CheckCalls()
	asserterror.Equal(len(arr), 1, t)

	err := checkMethodCallsInfo(arr[0], 2, 1)
	asserterror.EqualError(err, nil, t)

	reader.Read(nil)
	arr = reader.Mock.CheckCalls()
	asserterror.Equal(len(arr), 0, t)
}

func checkMethodCallsInfo(info mok.MethodCallsInfo, expectedCalls,
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
