package network

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// UDPClient represents a UDP client.
type UDPClient struct {
	serverAddr   string
	messageQueue queue.Queue
	conn         *net.UDPConn
}

// NewUDPClient creates a new UDP client.
func NewUDPClient(serverAddr string, messageQueue queue.Queue) *UDPClient {
	return &UDPClient{
		serverAddr: serverAddr,
	}
}

// Start starts the UDP client.
func (c *UDPClient) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	defer conn.Close()

	for {
		msg, err := ReceiveUDPMessage(conn)
		if err != nil {
			fmt.Printf("Failed to receive message from UDP connection: %v\n", err)
			continue
		}
		log.Debug("Received message from UDP server: %v", msg)

		// TODO: some messages might not make sense to queue (e.g. server pong)
		if err := c.messageQueue.Enqueue(msg); err != nil {
			log.Error("Failed to enqueue message: %v", err)
		}
	}
}

// SendMessage sends a message to the UDP server.
func (c *UDPClient) SendMessage(msg *messages.Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = c.conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write message to UDP connection: %v", err)
	}

	return nil
}

// receiveMessages continuously receives messages from the UDP server.
func ReceiveUDPMessage(conn *net.UDPConn) (*messages.Message, error) {
	buffer := make([]byte, messages.MessageBufferSize)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from UDP connection: %v", err)
	}

	msg := &messages.Message{}
	err = json.Unmarshal(buffer[:n], msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return msg, nil
}
