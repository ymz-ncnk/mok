package mok_test

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"

	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
	"github.com/ymz-ncnk/mok/testdata"
)

func TestMok(t *testing.T) {

	t.Run("Calls ok", func(t *testing.T) {
		var (
			wantNum1 = 1
			wantErr2 = errors.New("unexpeted param")
			reader   = testdata.NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) {
					if !bytes.Equal(p, []byte{1, 2, 3}) {
						return 0, errors.New("unexpeted param")
					}
					return wantNum1, nil
				},
			).RegisterRead(
				func(p []byte) (n int, err error) {
					if !bytes.Equal(p, []byte{4, 5}) {
						return 0, wantErr2
					}
					return 2, nil
				},
			)
		)
		n, err := reader.Read([]byte{1, 2, 3})
		asserterror.EqualError(err, nil, t)
		asserterror.Equal(n, wantNum1, t)

		_, err = reader.Read([]byte{})
		asserterror.EqualError(err, wantErr2, t)

		info := reader.Mock.CheckCalls()
		asserterror.EqualDeep(info, []mok.MethodCallsInfo{}, t)
	})

	t.Run("Unregister", func(t *testing.T) {
		var (
			wantErr = mok.NewUnknownMethodCallError("Reader", "Read")
			reader  = testdata.NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			)
		)
		reader.Mock.Unregister("Read")
		_, err := reader.Read([]byte{})
		asserterror.EqualError(err, wantErr, t)
	})

	t.Run("Unknown method call", func(t *testing.T) {
		var (
			wantErr = mok.NewUnknownMethodCallError("Reader", "ReadN")
			reader  = testdata.NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			)
		)
		result, err := reader.Mock.Call("ReadN", []byte{})
		asserterror.EqualDeep(result, []any(nil), t)
		asserterror.EqualError(err, wantErr, t)
		asserterror.Equal(wantErr.MockName(), "Reader", t)
		asserterror.Equal(wantErr.MethodName(), "ReadN", t)
	})

	t.Run("Unexpected call", func(t *testing.T) {
		var (
			wantErr = mok.NewUnexpectedMethodCallError("Reader", "Read")
			reader  = testdata.NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			)
		)
		reader.Mock.Call("Read", []byte{})
		result, err := reader.Mock.Call("Read", []byte{})
		asserterror.EqualDeep(result, []any(nil), t)
		asserterror.EqualDeep(err, wantErr, t)
		asserterror.Equal(wantErr.MockName(), "Reader", t)
		asserterror.Equal(wantErr.MethodName(), "Read", t)
	})

	t.Run("CheckCalls", func(t *testing.T) {
		var (
			reader = testdata.NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			).RegisterRead(
				func(p []byte) (n int, err error) { return 1, nil },
			)
		)
		reader.Read([]byte{})
		arr := reader.Mock.CheckCalls()
		asserterror.Equal(len(arr), 1, t)
		asserterror.EqualError(CheckMethodCallsInfo(arr[0], 2, 1), nil, t)
	})

	t.Run("Concurrent usage", func(t *testing.T) {
		var (
			wantNum1 = 10
			wantNum2 = 20
			wantNums = map[int]struct{}{wantNum1: {}, wantNum2: {}}
			wantErr  = mok.NewUnexpectedMethodCallError("Reader", "Read")
			wantErrs = map[string]struct{}{wantErr.Error(): {}}
			reader   = testdata.NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return wantNum1, nil },
			).RegisterRead(
				func(p []byte) (n int, err error) { return wantNum2, nil },
			)
			nums = make(chan int, 3)
			errs = make(chan error, 3)
			wg   = &sync.WaitGroup{}
		)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				n, err := reader.Read([]byte{})
				if err != nil {
					errs <- err
				} else {
					nums <- n
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(nums)
		close(errs)
		for n := range nums {
			delete(wantNums, n)
		}
		for err := range errs {
			delete(wantErrs, err.Error())
		}
		asserterror.Equal(len(wantNums), 0, t)
		asserterror.Equal(len(wantErrs), 0, t)

		arr := reader.Mock.CheckCalls()
		asserterror.EqualDeep(arr, []mok.MethodCallsInfo{}, t)
	})

	t.Run("Nil_param_caveat", func(t *testing.T) {
		var (
			wantErr error = nil
			writer        = testdata.NewWriterToMock().RegisterWriteTo(
				func(w io.Writer) (n int64, err error) { return 0, nil })
		)
		_, err := writer.WriteTo(nil)
		asserterror.EqualError(err, wantErr, t)
	})

	t.Run("Should be able to mock an interface with a variadic method",
		func(t *testing.T) {
			var (
				variadic = testdata.NewVariadicMock().RegisterProcess(
					func(args ...int) {},
				)
			)
			variadic.Process(1, 2, 3)
		})
}

func CheckMethodCallsInfo(info mok.MethodCallsInfo, expectedCalls,
	actualCalls int) error {
	if info.MockName != "Reader" {
		return errors.New("unexpected MockName")
	}
	if info.MethodName != "Read" {
		return errors.New("unexpected MethodName")
	}
	if info.ExpectedCalls != expectedCalls {
		return errors.New("unexpected ExpectedCalls")
	}
	if info.ActualCalls != actualCalls {
		return errors.New("unexpected ActualCalls")
	}
	return nil
}
