# Mok
With help of Mok you can mock any interface you want.

# Tests
Test coverage is about 90%.

# How to use
As an example, let's create a mock implementation of the `io.Reader` interface.
Create in your home directory a `foo` folder with the following structure:
```
foo/
 |‒‒‒reader_mock.go
 |‒‒‒reader_mock_test.go
```

__reader_mock.go__
```go
package foo

import "github.com/ymz-ncnk/mok"

func NewReaderMock() ReaderMock {
  return ReaderMock{mok.New("Reader")}
}

// ReaderMock is a mock implementation of the io.Reader. It simply uses
// mok.Mock as a delegate.
type ReaderMock struct {
  *mok.Mock
}


// RegisterRead registers a function as a single call to the Read() method.
func (reader ReaderMock) RegisterRead(
  fn func(p []byte) (n int, err error)) ReaderMock {
  reader.Register("Read", fn)
  return reader
}

// RegisterNRead registers a function as n calls to the Read() method.
func (reader ReaderMock) RegisterNRead(n int,
  fn func(p []byte) (n int, err error)) ReaderMock {
  reader.RegisterN("Read", n, fn)
  return reader
}

// UnregisterRead unregisters calls to the Read() method.
func (reader ReaderMock) UnregisterRead() ReaderMock {
  reader.Unregister("Read")
  return reader
}

func (reader ReaderMock) Read(p []byte) (n int, err error) {
  result, err := reader.Call("Read", p)
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
  "io"
  "testing"

  "github.com/ymz-ncnk/mok"
)

func TestSeveralCalls(t *testing.T) {
  // Here, we register several calls to the Read() method, and then call it 
  // several times as well.
  var (
    reader = func() ReaderMock {
      reader := NewReaderMock()
      // We must register all expected method calls. Each method call is just a 
      // function.
      reader.RegisterRead(func(p []byte) (n int, err error) {
        p[0] = 1
        return 1, nil
      }).RegisterRead(func(p []byte) (n int, err error) {
        p[0] = 2
        p[1] = 2
        return 2, nil
      })
      // If we want to register one function as several method calls, we use the 
      // RegisterN() method. This is especially useful for concurrent method 
      // calls.
      return reader.RegisterNRead(2, func(p []byte) (n int, err error) {
        return 0, io.EOF
      })
    }()
    b = make([]byte, 2)
  )

  // In total, we have registered 4 calls to the Read() method.

  // First call.
  n, err := reader.Read(b)
  if err != nil {
    panic(err)
  }
  // We expect to read 1 byte.
  if n != 1 {
    t.Errorf("unexpected n, want '%v', actual '%v'", 1, n)
  }
  // Here we could test err and b values ...

  // Second call.
  n, err = reader.Read(b)
  if err != nil {
    panic(err)
  }
  // We expect to read 2 bytes.
  if n != 2 {
    t.Errorf("unexpected n, want '%v', actual '%v'", 2, n)
  }

  // Third call.
  _, err = reader.Read(b)
  // We expect to receive io.EOF error.
  if err != io.EOF {
    t.Errorf("unexpected err, want '%v', actual '%v'", io.EOF, err)
  }

  // Forth call.
  _, err = reader.Read(b)
  // We expect to receive io.EOF error.
  if err != io.EOF {
    t.Errorf("unexpected err, want '%v', actual '%v'", io.EOF, err)
  }

  // If we call the Read() method again, we will get mok.UnexpectedMethodCallError.
  _, err = reader.Read(b)
  if err == nil {
    t.Error("unexpected nil error")
  } else {
    want := mok.NewUnexpectedMethodCallError("Reader", "Read")
    if err.Error() != want.Error() {
      t.Errorf("unexpected error, want '%v', actual '%v'", want, err)
    }
  }
}

func TestUnregisteredCall(t *testing.T) {
  // If we call a method without registered calls, we will get 
  // mok.UnknownMethodCallError.
  var (
    reader = NewReaderMock()
    b      []byte
  )
  _, err := reader.Read(b)
  if err == nil {
    t.Error("unexpected nil error")
  } else {
    want := mok.NewUnknownMethodCallError("Reader", "Read")
    if err.Error() != want.Error() {
      t.Errorf("unexpected error, want '%v', actual '%v'", want, err)
    }
  }
}

func TestCheckCallsFunction(t *testing.T) {
  // With the mok.CheckCalls() function, we can check whether all registered 
  // method calls have been used or not.
  var reader = func() ReaderMock {
      return NewReaderMock().RegisterRead(
        func(p []byte) (n int, err error) {
          p[0] = 1
          return 1, nil
        })
    }()
  infomap := mok.CheckCalls([]*mok.Mock{reader.Mock})
  if len(infomap) != 1 {
    t.Fatal("unexpected CheckCalls result")
  }
  if len(infomap) != 1 {
    t.Fatal("unexpected infomap length")
  }
  if len(infomap[0]) != 1 {
    t.Fatal("number of the MethodCallsInfo not equal to 1")
  }
  // test infomap[0] ...
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
To avoid this, we can pass `reflect.Value` to the `mok.Call()` function 
instead of `nil`. For example:
```go
// Mock implementation of the io.WriterTo.
type WriterToMock struct {
  *mok.Mock
}

func (writer WriterToMock) RegisterWriteTo(fn func(w io.Writer) (n int64,
  err error)) WriterToMock {
  writer.Register("WriteTo", fn)
  return writer
}

func (writer WriterToMock) WriteTo(w io.Writer) (n int64, err error) {
  // w param here may be nil, so we have to use reflect.Value.
  var wVal reflect.Value
  if w == nil {
    wVal = reflect.Zero(reflect.TypeOf((*io.Writer)(nil)).Elem())
  } else {
    wVal = reflect.ValueOf(w)
  }
  // Call() method can accept reflect.Value too.
  vals, err := writer.Call("WriteTo", wVal)
    if err != nil {
    return
  }
  n = vals[0].(int64)
  err, _ = vals[1].(error)
  return
}
```