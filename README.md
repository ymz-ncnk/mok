# Mok

[![Go Reference](https://pkg.go.dev/badge/github.com/ymz-ncnk/mok.svg)](https://pkg.go.dev/github.com/ymz-ncnk/mok)
[![GoReportCard](https://goreportcard.com/badge/ymz-ncnk/mok)](https://goreportcard.com/report/github.com/ymz-ncnk/mok)
[![codecov](https://codecov.io/gh/ymz-ncnk/mok/graph/badge.svg?token=8N19NXWZCQ)](https://codecov.io/gh/ymz-ncnk/mok)

With help of Mok you can mock any interface you want.

# Contents
- [Mok](#mok)
- [Contents](#contents)
- [Tests](#tests)
- [How to use](#how-to-use)
  - [Concurrent invocation](#concurrent-invocation)
- [Thread safety](#thread-safety)
- [Mock implementation caveats](#mock-implementation-caveats)

# Tests
Test coverage is about 90%.

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
  "errors"
  "io"
  "reflect"
  "sync"
  "testing"

  assert "github.com/ymz-ncnk/assert/testing_error"
  "github.com/ymz-ncnk/mok"
)

func TestSeveralCalls(t *testing.T) {
  // Here we register several calls to the Read() method, and then call it
  // several times as well. Each method call is just a function.
  // If we want to register one function several times, we can use the
  // RegisterN() method. This is especially useful when testing concurrent 
  // method invocations.
  var (
    reader = NewReaderMock().RegisterRead(func(p []byte) (n int, err error) {
      p[0] = 1
      return 1, nil
    }).RegisterRead(func(p []byte) (n int, err error) {
      p[0] = 2
      p[1] = 2
      return 2, nil
    }).RegisterNRead(2, func(p []byte) (n int, err error) {
      return 0, io.EOF
    })
    b = make([]byte, 2)
  )
  // In total, we have registered 4 calls to the Read() method.

  // First call.
  n, _ := reader.Read(b)
  // We expect to read 1 byte.
  assert.Equal(n, 1, t)
  // Here we could test err and b values ...

  // Second call.
  n, _ = reader.Read(b)
  // We expect to read 2 bytes.
  assert.Equal(n, 2, t)

  // Third call.
  _, err := reader.Read(b)
  // We expect to receive io.EOF error.
  assert.EqualError(err, io.EOF, t)

  // Forth call.
  _, err = reader.Read(b)
  // We expect to receive io.EOF error.
  assert.EqualError(err, io.EOF, t)

  // If we call the Read() method again, we will get mok.UnexpectedMethodCallError.
  _, err = reader.Read(b)
  assert.EqualError(err, mok.NewUnexpectedMethodCallError("Reader", "Read"), t)
}

func TestUnregisteredCall(t *testing.T) {
  // If we call a method without registered calls, we will get
  // mok.UnknownMethodCallError.
  var (
    reader = NewReaderMock()
    b      []byte
  )
  _, err := reader.Read(b)
  assert.EqualError(err, mok.NewUnknownMethodCallError("Reader", "Read"), t)
}

func TestCheckCallsFunction(t *testing.T) {
  // With mok.CheckCalls(), we can check whether all registered method calls
  // have been used.
  var (
    reader = NewReaderMock().RegisterRead(
      func(p []byte) (n int, err error) {
        p[0] = 1
        return 1, nil
      })
    mocks   = []*mok.Mock{reader.Mock}
    infomap = mok.CheckCalls(mocks)
  )
  // We have never called reader.Read(), so infomap is not empty.
  assert.Equal(len(infomap), 1, t)
  // In this case infomap[0] will contain []mok.MethodCallsInfo which
  // corresponds to the mocks[0] element.
  assert.EqualDeep(infomap[0], []mok.MethodCallsInfo{
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
## Concurrent invocation
To practice a little more, let's create an `io.Writer` mock:
```go
package foo

import "github.com/ymz-ncnk/mok"

type WriteFn func(p []byte) (n int, err error)

func NewWriterMock() WriterMock {
  return WriterMock{mok.New("Writer")}
}

type WriterMock struct {
  *mok.Mock
}

func (m WriterMock) RegisterNWrite(n int, fn WriteFn) WriterMock {
  m.RegisterN("Write", n, fn)
  return m
}

func (m WriterMock) Write(p []byte) (n int, err error) {
  result, err := m.Call("Write", p)
  if err != nil {
    return
  }
  n = result[0].(int)
  err, _ = result[1].(error)
  return
}

// WriteFn wrapper, performs the function only once.
type WriteFnWrapper struct {
  called bool
  fn     WriteFn
}

func (c *WriteFnWrapper) Write(p []byte) (n int, err error) {
  if c.called {
    err = errors.New("already called")
    return
  }
  c.called = true
  return c.fn(p)
}
```

And test concurrent invocation of its `Write` method:
```go
func TestConcurrentInvocation(t *testing.T) {
  var (
    p1 = []byte{1}
    p2 = []byte{2}

    fn1 = &WriteFnWrapper{fn: func(p []byte) (n int, err error) {
      return 1, nil
    }}
    fn2 = &WriteFnWrapper{fn: func(p []byte) (n int, err error) {
      return 2, nil
    }}

    // When preparing for several simultaneous calls, we cannot predict their
    // order, so we have to use RegisterN().
    writer = NewWriterMock().RegisterNWrite(2, func(p []byte) (n int, err error) {
      // The only thing we can do is to evaluate the input data.
      if reflect.DeepEqual(p, p1) {
        return fn1.Write(p)
      }
      if reflect.DeepEqual(p, p2) {
        return fn2.Write(p)
      }
      err = errors.New("unexpected input")
      return
    })
    wg = sync.WaitGroup{}
  )
  wg.Add(2)

  go func() {
    n, err := writer.Write(p1)
    assert.Equal(n, 1, t)
    assert.EqualError(err, nil, t)
    wg.Done()
  }()
  go func() {
    n, err := writer.Write(p2)
    assert.Equal(n, 2, t)
    assert.EqualError(err, nil, t)
    wg.Done()
  }()

  wg.Wait()
}
```

# Thread safety
The mock implementation is fully thread-safe. You can register, unregister, 
call methods, and check the number of calls simultaneously.

# Mock implementation caveats
Calling the `mok.Call()` method with a `nil` parameter can cause a panic like:
```
  panic: reflect: Call using zero Value argument
  ...
```
To avoid this, we can use `mok.SafeVal()` function:
```go
type WriteToFn func(w io.Writer) (n int64, err error)

// Mock implementation of the io.WriterTo.
type WriterToMock struct {
  *mok.Mock
}

func (m WriterToMock) RegisterWriteTo(fn WriteToFn) WriterToMock {
  m.Register("WriteTo", fn)
  return m
}

func (m WriterToMock) WriteTo(w io.Writer) (n int64, err error) {
  // w param here may be nil, so we have to use mok.SafeVal() function.
  vals, err := m.Call("WriteTo", mok.SafeVal[io.Writer](w))
    if err != nil {
    return
  }
  n = vals[0].(int64)
  err, _ = vals[1].(error)
  return
}
```