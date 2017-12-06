package rpctypes

import "github.com/SeerUK/i3x3/pkg/proto"

// Message is a message container that provides the structure to have more of a request / response
// cycle. Once the message has been processed, a response can be issued by passing the result (an
// error, or nil) to the response channel.
type Message struct {
	// Command is a command that comes from an RPC client.
	Command proto.DaemonCommand
	// ResponseCh is a channel to send a response down. The response may simply be nil, indicating
	// success. An error sent down this channel will likely be sent to the client (i3x3ctl).
	ResponseCh chan<- error
}

// NewMessage creates a new RPC message, and returns a channel that a response should be passed to.
func NewMessage(command *proto.DaemonCommand) (Message, chan error) {
	responseCh := make(chan error)
	message := Message{
		ResponseCh: responseCh,
		Command:    *command,
	}

	return message, responseCh
}
