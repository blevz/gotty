package webtty

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
)

// WebTTY bridges a PTY upstream and its PTY downstream.
// To support text-based streams and side channel commands such as
// terminal resizing, WebTTY uses an original protocol.
type WebTTY struct {
	// PTY downstream, which is usually a connection to browser
	downstreamReader DownstreamReader
	downstreamWriter DownstreamWriter
	// PTY upstream, usually a local tty
	upstream Upstream

	windowTitle []byte
	permitWrite bool
	columns     int
	rows        int
	reconnect   int // in seconds
	masterPrefs []byte

	bufferSize int
	writeMutex sync.Mutex
}

// New creates a new instance of WebTTY.
// downstream is a connection to the PTY downstream,
// typically it's a websocket connection to a client.
// upstream is usually a local command with a PTY.
func New(options ...Option) (*WebTTY, error) {
	wt := &WebTTY{
		permitWrite: false,
		columns:     0,
		rows:        0,

		bufferSize: 1024,
	}

	for _, option := range options {
		option(wt)
	}

	return wt, nil
}

// Run starts the main process of the WebTTY.
// This method blocks until the context is canceled.
// Note that the downstream and upstream are left intact even
// after the context is canceled. Closing them is caller's
// responsibility.
// If the connection to one end gets closed, returns ErrUpstreamClosed or ErrDownstreamClosed.
func (wt *WebTTY) Run(ctx context.Context) error {
	err := wt.sendInitializeMessage()
	if err != nil {
		return errors.Wrapf(err, "failed to send initializing message")
	}

	errs := make(chan error, 2)

	go func() {
		errs <- func() error {
			buffer := make([]byte, wt.bufferSize)
			for {
				n, err := wt.upstream.Read(buffer)
				if err != nil {
					return ErrUpstreamClosed
				}

				err = wt.handleUpstreamReadEvent(buffer[:n])
				if err != nil {
					return err
				}
			}
		}()
	}()

	go func() {
		errs <- func() error {
			for {
				reqType, data, err := wt.downstreamReader.ReadMessage()
				if err != nil {
					return ErrDownstreamClosed
				}
				err = wt.handleMasterReadEvent(reqType, data)
				if err != nil {
					return err
				}
			}
		}()
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errs:
	}

	return err
}

func (wt *WebTTY) sendInitializeMessage() error {
	err := wt.masterWrite(SetWindowTitle, wt.windowTitle)
	if err != nil {
		return errors.Wrapf(err, "failed to send window title")
	}

	if wt.reconnect > 0 {
		reconnect, _ := json.Marshal(wt.reconnect)
		err := wt.masterWrite(SetReconnect, reconnect)
		if err != nil {
			return errors.Wrapf(err, "failed to set reconnect")
		}
	}

	if wt.masterPrefs != nil {
		err := wt.masterWrite(SetPreferences, wt.masterPrefs)
		if err != nil {
			return errors.Wrapf(err, "failed to set preferences")
		}
	}

	return nil
}

func (wt *WebTTY) handleUpstreamReadEvent(data []byte) error {
	safeMessage := base64.StdEncoding.EncodeToString(data)
	err := wt.masterWrite(Output, []byte(safeMessage))
	if err != nil {
		return errors.Wrapf(err, "failed to send message to master")
	}

	return nil
}

func (wt *WebTTY) masterWrite(r ResponseType, data []byte) error {
	wt.writeMutex.Lock()
	defer wt.writeMutex.Unlock()

	err := wt.downstreamWriter.WriteMessage(r, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write to master")
	}

	return nil
}

func (wt *WebTTY) handleMasterReadEvent(r RequestType, data []byte) error {
	switch r {
	case Input:
		if err := wt.handleInput(data); err != nil {
			return err
		}

	case Ping:
		err := wt.handlePong()
		if err != nil {
			return err
		}

	case ResizeTerminal:
		var args argResizeTerminal
		err := json.Unmarshal(data, &args)
		if err != nil {
			return errors.Wrapf(err, "received malformed data for terminal resize")
		}
		err = wt.handleResize(args)
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown message type `%c`", data[0])
	}

	return nil
}

func (wt *WebTTY) handleResize(args argResizeTerminal) error {
	rows := wt.rows
	if rows == 0 {
		rows = int(args.Rows)
	}

	columns := wt.columns
	if columns == 0 {
		columns = int(args.Columns)
	}

	wt.upstream.ResizeTerminal(columns, rows)
	return nil
}

func (wt *WebTTY) handlePong() error {
	err := wt.masterWrite(Pong, []byte{})
	if err != nil {
		return errors.Wrapf(err, "failed to return Pong message to downstream")
	}
	return nil
}

func (wt *WebTTY) handleInput(data []byte) error {
	if !wt.permitWrite {
		return nil
	}
	_, err := wt.upstream.Write(data)
	if err != nil {
		return errors.Wrapf(err, "failed to write received data to upstream")
	}
	return nil
}

type argResizeTerminal struct {
	Columns float64
	Rows    float64
}
