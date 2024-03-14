package network

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// TCPClient represents a TCP client.
type TCPClient struct {
	serverAddr   string
	messageQueue queue.Queue
	conn         net.Conn
}

// NewTCPClient creates a new TCP client.
func NewTCPClient(serverAddr string, messageQueue queue.Queue) *TCPClient {
	return &TCPClient{
		serverAddr:   serverAddr,
		messageQueue: messageQueue,
	}
}

func (c *TCPClient) Connect(clientIDChan chan uint32) error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()
	c.conn = conn

	for {
		msg, err := ReceiveTCPMessage(conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosed); ok {
				return err
			}
			log.Error("Failed to receive message from TCP connection: %v", err)
			continue
		}
		log.Trace("Received message from TCP server of type %s", msg.Type)

		switch msg.Type {
		case messages.MessageTypeServerAssignID:
			assignID := &messages.AssignID{}
			err := json.Unmarshal(msg.Payload, assignID)
			if err != nil {
				log.Error("Failed to deserialize server assign ID message: %v", err)
				continue
			}
			// write the client ID back to the manager
			clientIDChan <- assignID.ClientID
		default:
			log.Warn("Received unexpected message type from TCP server: %s", msg.Type)
			continue
		}
	}
}

// SendMessage sends a message to the TCP server.
func (c *TCPClient) SendMessage(msg *messages.Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = c.conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write message to TCP connection: %v", err)
	}

	return nil
}

// ErrConnectionClosed is returned when the TCP connection is closed
type ErrConnectionClosed struct{}

func (e *ErrConnectionClosed) Error() string {
	return "connection closed"
}

// ReceiveTCPMessage receives a message from the TCP server.
func ReceiveTCPMessage(conn net.Conn) (*messages.Message, error) {
	jsonData := make([]byte, messages.MessageBufferSize)
	n, err := conn.Read(jsonData)
	if err != nil {
		if err.Error() == "EOF" {
			return nil, &ErrConnectionClosed{}
		}
		return nil, fmt.Errorf("failed to read message from TCP connection: %v", err)
	}

	msg := &messages.Message{}
	err = json.Unmarshal(jsonData[:n], msg)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
