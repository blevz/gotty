package webtty

import (
	"io"
)

// Downstream represents a PTY sending input and receiving output from the Upstream, usually it's a websocket connection.
type DownstreamReader io.Reader
type DownstreamWriter io.Writer
type Downstream io.ReadWriter
