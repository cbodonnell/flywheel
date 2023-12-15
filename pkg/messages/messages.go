package messages

import (
	"encoding/json"
	"fmt"
	"net"
)

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

// WriteMessageToTCP writes a Message to a TCP connection
func WriteMessageToTCP(conn net.Conn, msg Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write message to TCP connection: %v", err)
	}

	return nil
}

// ReadMessageFromTCP reads a Message from a TCP connection
func ReadMessageFromTCP(conn net.Conn) (Message, error) {
	var msg Message

	jsonData := make([]byte, MessageBufferSize) // Adjust buffer size based on your needs
	n, err := conn.Read(jsonData)
	if err != nil {
		return msg, fmt.Errorf("failed to read message from TCP connection: %v", err)
	}

	err = json.Unmarshal(jsonData[:n], &msg)
	if err != nil {
		return msg, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}

// WriteMessageToUDP writes a Message to a UDP connection
func WriteMessageToUDP(conn *net.UDPConn, msg Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write message to UDP connection: %v", err)
	}

	return nil
}

// ReadMessageFromUDP reads a Message from a UDP connection
func ReadMessageFromUDP(conn *net.UDPConn) (Message, *net.UDPAddr, error) {
	var msg Message

	jsonData := make([]byte, MessageBufferSize) // Adjust buffer size based on your needs
	n, addr, err := conn.ReadFromUDP(jsonData)
	if err != nil {
		return msg, nil, fmt.Errorf("failed to read message from UDP connection: %v", err)
	}

	err = json.Unmarshal(jsonData[:n], &msg)
	if err != nil {
		return msg, nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, addr, nil
}
