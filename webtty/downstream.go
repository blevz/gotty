package webtty

import (
	"io"
)

// Downstream represents a PTY sending input and receiving output from the Upstream, usually it's a websocket connection.
type DownstreamReader interface {
	ReadMessage() (RequestType, []byte, error)
}
type DownstreamWriter interface {
	WriteMessage(ResponseType, []byte) error
}
type Downstream interface {
	DownstreamReader
	DownstreamWriter
}

type CoalescingDownstreamWriter struct {
	Writers []DownstreamWriter
}

type DownstreamWriterWrapper struct {
	Writer io.Writer
}

type DownstreamReaderCapture struct {
	Reader io.Reader
}

type DownstreamReaderWrapper struct {
	Reader io.Reader
}

func (w CoalescingDownstreamWriter) WriteMessage(r ResponseType, data []byte) error {
	for _, writer := range w.Writers {
		err := writer.WriteMessage(r, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d DownstreamReaderWrapper) ReadMessage() (RequestType, []byte, error) {
	buffer := make([]byte, 1024)
	n, err := d.Reader.Read(buffer)
	if err != nil {
		return UnknownInput, nil, ErrDownstreamClosed
	}
	t := RequestType(buffer[0])
	if t == Ping || n < 2 {
		return t, nil, nil
	}
	return t, buffer[1:n], nil
}

func (d DownstreamWriterWrapper) WriteMessage(r ResponseType, data []byte) error {
	message := []byte(string(r))
	message = append(message, data...)
	_, err := d.Writer.Write(message)
	return err
}
