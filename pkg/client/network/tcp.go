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

func (c *TCPClient) Start() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()
	c.conn = conn

	for {
		msg, err := ReceiveTCPMessage(conn)
		if err != nil {
			log.Error("Failed to receive message from TCP connection: %v", err)
			continue
		}
		log.Debug("Received message from TCP server: %v", msg)

		// TODO: some messages might not make sense to queue (e.g. server pong)
		if err := c.messageQueue.Enqueue(msg); err != nil {
			log.Error("Failed to enqueue message: %v", err)
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

// ReceiveTCPMessage receives a message from the TCP server.
func ReceiveTCPMessage(conn net.Conn) (*messages.Message, error) {
	jsonData := make([]byte, messages.MessageBufferSize)
	n, err := conn.Read(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from TCP connection: %v", err)
	}

	msg := &messages.Message{}
	err = json.Unmarshal(jsonData[:n], msg)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
