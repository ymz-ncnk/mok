package testdata

import "github.com/ymz-ncnk/mok"

type ReadFn func(p []byte) (n int, err error)

func NewReaderMock() ReaderMock {
	return ReaderMock{mok.New("Reader")}
}

// Mock implementation of the io.Reader interface.
type ReaderMock struct {
	*mok.Mock
}

func (r ReaderMock) RegisterRead(fn ReadFn) ReaderMock {
	r.Register("Read", fn)
	return r
}

func (r ReaderMock) Read(p []byte) (n int, err error) {
	result, err := r.Call("Read", p)
	if err != nil {
		return 0, err
	}
	n = result[0].(int)
	err, _ = result[1].(error)
	return
}
