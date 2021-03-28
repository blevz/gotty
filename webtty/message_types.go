package webtty

// Protocols defines the name of this protocol,
// which is supposed to be used to the subprotocol of Websockt streams.
var Protocols = []string{"webtty"}

type RequestType rune

const (
	// Unknown message type, maybe sent by a bug
	UnknownInput RequestType = '0'
	// User input typically from a keyboard
	Input RequestType = '1'
	// Ping to the server
	Ping RequestType = '2'
	// Notify that the browser size has been changed
	ResizeTerminal RequestType = '3'
)

type ResponseType rune

const (
	// Unknown message type, maybe set by a bug
	UnknownOutput ResponseType = '0'
	// Normal output to the terminal
	Output ResponseType = '1'
	// Pong to the browser
	Pong ResponseType = '2'
	// Set window title of the terminal
	SetWindowTitle ResponseType = '3'
	// Set terminal preference
	SetPreferences ResponseType = '4'
	// Make terminal to reconnect
	SetReconnect ResponseType = '5'
)
