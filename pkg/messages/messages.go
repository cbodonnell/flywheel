package messages

const (
	// MessageBufferSize represents the maximum size of a message
	MessageBufferSize = 1024
)

// Message types
const (
	MessageTypeClientPing   = "clientPing"
	MessageTypeServerPong   = "serverPong"
	MessageTypeClientUpdate = "clientUpdate"
	MessageTypeServerUpdate = "serverUpdate"
)

// Message represents a generic message for serialization/deserialization
type Message struct {
	ClientID uint32      `json:"clientID"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload"`
}
