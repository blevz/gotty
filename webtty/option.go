package webtty

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Option is an option for WebTTY.
type Option func(*WebTTY) error

// WithUpstream sets the upstream tty, usually a local a local command w/ a PTY.
func WithUpstream(upstream Upstream) Option {
	return func(wt *WebTTY) error {
		wt.upstream = upstream
		return nil
	}
}

func WithDownstream(downstream Downstream) Option {
	return func(wt *WebTTY) error {
		wt.downstreamReader = downstream
		wt.downstreamWriter = downstream
		return nil
	}
}

func WithDownstreamReader(downstreamReader DownstreamReader) Option {
	return func(wt *WebTTY) error {
		wt.downstreamReader = downstreamReader
		return nil
	}
}

func WithDownstreamWriter(downstreamWriter DownstreamWriter) Option {
	return func(wt *WebTTY) error {
		wt.downstreamWriter = downstreamWriter
		return nil
	}
}

// WithPermitWrite sets a WebTTY to accept input from upstreams.
func WithPermitWrite() Option {
	return func(wt *WebTTY) error {
		wt.permitWrite = true
		return nil
	}
}

// WithFixedColumns sets a fixed width to TTY downstreams.
func WithFixedColumns(columns int) Option {
	return func(wt *WebTTY) error {
		wt.columns = columns
		return nil
	}
}

// WithFixedRows sets a fixed height to TTY downstreams.
func WithFixedRows(rows int) Option {
	return func(wt *WebTTY) error {
		wt.rows = rows
		return nil
	}
}

// WithWindowTitle sets the default window title of the session
func WithWindowTitle(windowTitle []byte) Option {
	return func(wt *WebTTY) error {
		wt.windowTitle = windowTitle
		return nil
	}
}

// WithReconnect enables reconnection on the master side.
func WithReconnect(timeInSeconds int) Option {
	return func(wt *WebTTY) error {
		wt.reconnect = timeInSeconds
		return nil
	}
}

// WithMasterPreferences sets an optional configuration of master.
func WithMasterPreferences(preferences interface{}) Option {
	return func(wt *WebTTY) error {
		prefs, err := json.Marshal(preferences)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal preferences as JSON")
		}
		wt.masterPrefs = prefs
		return nil
	}
}
