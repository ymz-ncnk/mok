package mok

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"sync"
	"testing"
)

func NewReaderMock() ReaderMock {
	return ReaderMock{New("Reader")}
}

// Mock implementation of the io.Reader interface.
type ReaderMock struct {
	*Mock
}

func (reader ReaderMock) RegisterRead(
	fn func(p []byte) (n int, err error)) ReaderMock {
	reader.Register("Read", fn)
	return reader
}

func (reader ReaderMock) Read(p []byte) (n int, err error) {
	result, err := reader.Call("Read", p)
	if err != nil {
		return 0, err
	}
	if result[1] != nil {
		err = result[1].(error)
	}
	return result[0].(int), err
}

// -----------------------------------------------------------------------------
func NeWriterToMock() WriterToMock {
	return WriterToMock{New("WriterTo")}
}

// Mock implementation of the io.WriterTo interface.
type WriterToMock struct {
	*Mock
}

func (writer WriterToMock) RegisterWriteTo(fn func(w io.Writer) (n int64,
	err error)) WriterToMock {
	writer.Register("WriteTo", fn)
	return writer
}

func (writer WriterToMock) WriteTo(w io.Writer) (n int64, err error) {
	var wVal reflect.Value
	if w == nil {
		wVal = reflect.Zero(reflect.TypeOf((*io.Writer)(nil)).Elem())
	} else {
		wVal = reflect.ValueOf(w)
	}
	vals, err := writer.Call("WriteTo", wVal)
	if err != nil {
		return
	}
	n = vals[0].(int64)
	err, _ = vals[1].(error)
	return
}

// -----------------------------------------------------------------------------
func TestMock(t *testing.T) {
	t.Run("Calls ok", func(t *testing.T) {
		var (
			wantNum1 = 1
			wantNum2 = errors.New("unexpeted param")
			reader   = NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) {
					if !bytes.Equal(p, []byte{1, 2, 3}) {
						return 0, errors.New("unexpeted param")
					}
					return wantNum1, nil
				},
			).RegisterRead(
				func(p []byte) (n int, err error) {
					if !bytes.Equal(p, []byte{4, 5}) {
						return 0, wantNum2
					}
					return 2, nil
				},
			)
		)

		n, err := reader.Read([]byte{1, 2, 3})
		if err != nil {
			t.Error(err)
		}
		if n != wantNum1 {
			t.Error("unexpected result")
		}
		_, err = reader.Read([]byte{})
		if err != wantNum2 {
			t.Error("unexpected error")
		}
		info := reader.CheckCalls()
		if len(info) != 0 {
			t.Error("unexpected CheckCalls result")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		var (
			want   = UnknownMethodCallError{"Reader", "Read"}
			reader = NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			)
		)

		reader.Unregister("Read")
		_, err := reader.Read([]byte{})
		if err.Error() != want.Error() {
			t.Error("unexpected error")
		}
	})

	t.Run("Unknown method call", func(t *testing.T) {
		var (
			want   = UnknownMethodCallError{"Reader", "ReadN"}
			reader = NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			)
		)

		result, err := reader.Call("ReadN", []byte{})
		if result != nil {
			t.Error("unexpected result")
		}
		if err.Error() != want.Error() {
			t.Error("unexpected error")
		}
		if want.MockName() != "Reader" {
			t.Error("unexpected MockName")
		}
		if want.MethodName() != "ReadN" {
			t.Error("unexpected MethodName")
		}
	})

	t.Run("Unexpected call", func(t *testing.T) {
		var (
			want   = UnexpectedMethodCallError{"Reader", "Read"}
			reader = NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			)
		)

		reader.Call("Read", []byte{})
		result, err := reader.Call("Read", []byte{})
		if result != nil {
			t.Error("unexpected result")
		}
		if err.Error() != want.Error() {
			t.Error("unexpected error")
		}
		if want.MockName() != "Reader" {
			t.Error("unexpected MockName")
		}
		if want.MethodName() != "Read" {
			t.Error("unexpected MethodName")
		}
	})

	t.Run("CheckCalls", func(t *testing.T) {
		var (
			reader = NewReaderMock().RegisterRead(
				func(p []byte) (n int, err error) { return 0, nil },
			).RegisterRead(
				func(p []byte) (n int, err error) { return 1, nil },
			)
		)

		reader.Read([]byte{})
		arr := reader.CheckCalls()
		if len(arr) != 1 {
			t.Error("unexpected CheckCalls result")
		}
		err := CheckMethodCallsInfo(arr[0], 2, 1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Concurrent usage", func(t *testing.T) {
		var (
			wantNum1 = 10
			wantNum2 = 20
			wantNums = map[int]struct{}{wantNum1: {}, wantNum2: {}}
			wantErr  = UnexpectedMethodCallError{"Reader", "Read"}
			wantErrs = map[string]struct{}{wantErr.Error(): {}}
			reader   = NewReaderMock().RegisterRead(
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
		if len(wantNums) != 0 {
			t.Error("unexpected num")
		}
		if len(wantErrs) != 0 {
			t.Error("unexpected err")
		}
		arr := reader.CheckCalls()
		if len(arr) != 0 {
			t.Error("unexpected CheckCalls result")
		}
	})

	t.Run("Nil_param_caveat", func(t *testing.T) {
		var (
			writer = NeWriterToMock().RegisterWriteTo(
				func(w io.Writer) (n int64, err error) { return 0, nil })
		)
		_, err := writer.WriteTo(nil)
		if err != nil {
			t.Error(err)
		}
	})
}

func CheckMethodCallsInfo(info MethodCallsInfo, expectedCalls,
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
