package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// TCPClient represents a TCP client.
type TCPClient struct {
	serverAddr     string
	messageQueue   queue.Queue
	clientIDChan   chan<- uint32
	serverTimeChan chan<- *messages.ServerSyncTime
	conn           net.Conn
}

// NewTCPClient creates a new TCP client.
func NewTCPClient(serverAddr string, messageQueue queue.Queue, clientIDChan chan<- uint32, serverTimeChan chan<- *messages.ServerSyncTime) *TCPClient {
	return &TCPClient{
		serverAddr:     serverAddr,
		messageQueue:   messageQueue,
		clientIDChan:   clientIDChan,
		serverTimeChan: serverTimeChan,
	}
}

func (c *TCPClient) Connect() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	return nil
}

func (c *TCPClient) HandleMessages(ctx context.Context) error {
	defer c.conn.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := ReceiveTCPMessage(c.conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosedByServer); ok {
				return err
			} else if _, ok := err.(*ErrConnectionClosedByClient); ok {
				log.Info("TCP connection closed by client")
				return nil
			}
			log.Error("Failed to receive message from TCP connection: %v", err)
			continue
		}
		log.Trace("Received message from TCP server of type %s", msg.Type)

		switch msg.Type {
		case messages.MessageTypeServerLoginSuccess:
			assignID := &messages.ServerLoginSuccess{}
			err := json.Unmarshal(msg.Payload, assignID)
			if err != nil {
				log.Error("Failed to deserialize server assign ID message: %v", err)
				continue
			}
			// write the client ID back to the manager
			c.clientIDChan <- assignID.ClientID
		case messages.MessageTypeServerSyncTime:
			serverSyncTime := &messages.ServerSyncTime{}
			err := json.Unmarshal(msg.Payload, serverSyncTime)
			if err != nil {
				log.Error("Failed to deserialize server sync time message: %v", err)
				continue
			}
			c.serverTimeChan <- serverSyncTime
		case messages.MessageTypeServerPlayerConnect,
			messages.MessageTypeServerPlayerDisconnect,
			messages.MessageTypeServerNPCHit,
			messages.MessageTypeServerNPCKill:
			if err := c.messageQueue.Enqueue(msg); err != nil {
				log.Error("Failed to enqueue message: %v", err)
			}
		default:
			log.Warn("Received unexpected message type from TCP server: %s", msg.Type)
			continue
		}
	}
}

// Close closes the TCP connection.
func (c *TCPClient) Close() error {
	if c.conn == nil {
		log.Warn("TCP connection is already closed")
		return nil
	}
	return c.conn.Close()
}

// SendMessage sends a message to the TCP server.
func (c *TCPClient) SendMessage(msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = c.conn.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message to TCP connection: %v", err)
	}

	return nil
}

// ReceiveTCPMessage receives a message from the TCP server.
func ReceiveTCPMessage(conn net.Conn) (*messages.Message, error) {
	buf := make([]byte, messages.UDPMessageBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return nil, &ErrConnectionClosedByServer{}
		}
		if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
			return nil, &ErrConnectionClosedByClient{}
		}
		return nil, fmt.Errorf("failed to read message from TCP connection: %v", err)
	}

	msg, err := messages.DeserializeMessage(buf[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
