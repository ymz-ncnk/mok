# Mok

[![Go Reference](https://pkg.go.dev/badge/github.com/ymz-ncnk/mok.svg)](https://pkg.go.dev/github.com/ymz-ncnk/mok)
[![GoReportCard](https://goreportcard.com/badge/ymz-ncnk/mok)](https://goreportcard.com/report/github.com/ymz-ncnk/mok)
[![codecov](https://codecov.io/gh/ymz-ncnk/mok/graph/badge.svg?token=8N19NXWZCQ)](https://codecov.io/gh/ymz-ncnk/mok)

With help of Mok you can mock any interface you want.

# Contents
- [Mok](#mok)
- [Contents](#contents)
- [How to use](#how-to-use)
- [Concurrent Invocations](#concurrent-invocations)
- [Thread Safety](#thread-safety)

# How to use
As an example, let's mock the `io.Reader` interface. Create a `foo` folder 
with the following structure:
```
foo/
 |‒‒‒reader_mock.go
 |‒‒‒reader_mock_test.go
```

__reader_mock.go__
```go
package foo

import "github.com/ymz-ncnk/mok"

type ReadFn func(p []byte) (n int, err error)

func NewReaderMock() ReaderMock {
  return ReaderMock{mok.New("Reader")}
}

// ReaderMock is a mock implementation of the io.Reader. It simply uses
// mok.Mock as a delegate.
type ReaderMock struct {
  *mok.Mock
}

// RegisterRead registers a function as a single call to the Read() method.
func (m ReaderMock) RegisterRead(fn ReadFn) ReaderMock {
  m.Register("Read", fn)
  return m
}

// RegisterNRead registers a function as n calls to the Read() method.
func (m ReaderMock) RegisterNRead(n int, fn ReadFn) ReaderMock {
  m.RegisterN("Read", n, fn)
  return m
}

// UnregisterRead unregisters calls to the Read() method.
func (m ReaderMock) UnregisterRead() ReaderMock {
  m.Unregister("Read")
  return m
}

func (m ReaderMock) Read(p []byte) (n int, err error) {
  result, err := m.Call("Read", p)
  // Err here could be one of mok.UnexpectedMethodCallError or
  // mok.UnknownMethodCallError.
  if err != nil {
    return
  }
  // To safely call m.Call() with a possibly nil value, without triggering:
  // 
  //   panic: reflect: Call using zero Value argument
  //
  // , use the mok.SafeVal() helper. For example:
  //
  //   ... := m.Call("WriteTo", mok.SafeVal[io.Writer](w))

  n = result[0].(int)
  err, _ = result[1].(error)
  return
}
```
Run from the command line:
```bash
$ cd ~/foo
$ go mod init foo
$ go get github.com/ymz-ncnk/mok
```
Now, to see how the mock implementation works, let's test it.

__reader_mock_test.go__
```go
package foo

import (
	"io"
	"testing"

	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestSeveralCalls(t *testing.T) {
	// Here we register several calls to the Read() method, and then call it
	// several times as well. Each method call is just a function.
	// If we want to register one function several times, we can use the
	// RegisterN() method. This is especially useful when testing concurrent
	// method invocations.
	var (
		reader = NewReaderMock().RegisterRead(
			func(p []byte) (n int, err error) {
				p[0] = 1
				return 1, nil
			},
		).RegisterRead(
			func(p []byte) (n int, err error) {
				p[0] = 2
				p[1] = 2
				return 2, nil
			},
		).RegisterNRead(2,
			func(p []byte) (n int, err error) {
				return 0, io.EOF
			},
		)
		b = make([]byte, 2)
	)
	// In total, we have registered 4 calls to the Read() method.

	// First call.
	n, _ := reader.Read(b)
	// We expect to read 1 byte.
	asserterror.Equal(n, 1, t)
	// Here we could test err and b values ...

	// Second call.
	n, _ = reader.Read(b)
	// We expect to read 2 bytes.
	asserterror.Equal(n, 2, t)

	// Third call.
	_, err := reader.Read(b)
	// We expect to receive io.EOF error.
	asserterror.EqualError(err, io.EOF, t)

	// Forth call.
	_, err = reader.Read(b)
	// We expect to receive io.EOF error.
	asserterror.EqualError(err, io.EOF, t)

	// If we call the Read() method again, we will get mok.UnexpectedMethodCallError.
	_, err = reader.Read(b)
	asserterror.EqualError(err, mok.NewUnexpectedMethodCallError("Reader", "Read"), t)
}

func TestUnregisteredCall(t *testing.T) {
	// If we call a method without registered calls, we will get
	// mok.UnknownMethodCallError.
	var (
		reader = NewReaderMock()
		b      []byte
	)
	_, err := reader.Read(b)
	asserterror.EqualError(err, mok.NewUnknownMethodCallError("Reader", "Read"), t)
}

func TestCheckCallsFunction(t *testing.T) {
	// With mok.CheckCalls(), we can check whether all registered method calls
	// have been used.
	var (
		reader = NewReaderMock().RegisterRead(
			func(p []byte) (n int, err error) {
				p[0] = 1
				return 1, nil
			},
		)
		mocks   = []*mok.Mock{reader.Mock}
		infomap = mok.CheckCalls(mocks)
	)
	// We have never called reader.Read(), so infomap is not empty.
	asserterror.Equal(len(infomap), 1, t)
	// In this case infomap[0] will contain []mok.MethodCallsInfo which
	// corresponds to the mocks[0] element.
	asserterror.EqualDeep(infomap[0], []mok.MethodCallsInfo{
		{
			MockName:      "Reader",
			MethodName:    "Read",
			ExpectedCalls: 1,
			ActualCalls:   0,
		},
	}, t)
	// len(infomap) == 0 if all registered method calls have been used.
}
```
# Concurrent Invocations
Mocking a method for several concurrent invocations, using, for example:
```go
mock.RegisterWrite(
  func() { ... },
).RegisterWrite(
  func() { ... },
)
```
will not work, because the invocation order is not guaranteed. Instead, use 
`RegisterN()`:
```go
mock.RegisterNWrite(2, 
  func() { ... },
)
```
which registers a single function for two method calls.

# Thread Safety
The mock implementation is fully thread-safe. You can register, unregister, 
call methods, and check the number of calls simultaneously.
