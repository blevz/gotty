package utils

import "io"

type NewLineWriter struct {
	Writer io.Writer
}

func (w NewLineWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(append(p, []byte("\n")...))
}
