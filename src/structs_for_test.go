package main

type FakeWriter struct {
	buf []byte
}

func (w *FakeWriter) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return n, err
}

func (w *FakeWriter) CleanBuf() {
	w.buf = w.buf[:0]
}

func (w *FakeWriter) ToStringAndClean() string {
	s := string(w.buf)
	w.CleanBuf()

	return s
}
