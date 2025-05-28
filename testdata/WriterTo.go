package testdata

import (
	"io"

	"github.com/ymz-ncnk/mok"
)

type WriteToFn func(w io.Writer) (n int64, err error)

func NewWriterToMock() WriterToMock {
	return WriterToMock{mok.New("WriterTo")}
}

// Mock implementation of the io.WriterTo interface.
type WriterToMock struct {
	*mok.Mock
}

func (writer WriterToMock) RegisterWriteTo(fn WriteToFn) WriterToMock {
	writer.Register("WriteTo", fn)
	return writer
}

func (writer WriterToMock) WriteTo(w io.Writer) (n int64, err error) {
	vals, err := writer.Call("WriteTo", mok.SafeVal[io.Writer](w))
	if err != nil {
		return
	}
	n = vals[0].(int64)
	err, _ = vals[1].(error)
	return
}
