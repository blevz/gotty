package utils

import (
	"fmt"
	"io"
	"time"
)

type NewLineWriter struct {
	Writer io.Writer
}

func (w NewLineWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(append(p, []byte("\n")...))
}

type TimestampWriter struct {
	Writer io.Writer
}

func (w TimestampWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(fmt.Sprintf("%d ", time.Now().UnixNano())), p...))
}

type TaggedWriter struct {
	Tag    string
	Writer io.Writer
}

func (w TaggedWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(w.Tag+" "), p...))
}
