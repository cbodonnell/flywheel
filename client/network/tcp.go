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
	loginErrChan   chan<- error
	serverTimeChan chan<- *messages.ServerSyncTime
	conn           net.Conn
}

// NewTCPClient creates a new TCP client.
func NewTCPClient(serverAddr string, messageQueue queue.Queue, clientIDChan chan<- uint32, loginErrChan chan<- error, serverTimeChan chan<- *messages.ServerSyncTime) *TCPClient {
	return &TCPClient{
		serverAddr:     serverAddr,
		messageQueue:   messageQueue,
		clientIDChan:   clientIDChan,
		loginErrChan:   loginErrChan,
		serverTimeChan: serverTimeChan,
	}
}

func (c *TCPClient) Connect() error {
	log.Info("Connecting to TCP server at %s", c.serverAddr)
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

		b, err := ReadFromTCP(c.conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosedByServer); ok {
				return err
			} else if _, ok := err.(*ErrConnectionClosedByClient); ok {
				log.Info("TCP connection closed by client")
				return nil
			}
			log.Error("Failed to read from TCP connection: %v", err)
			continue
		}
		go func() {
			if err := c.handleMessage(b); err != nil {
				log.Error("Failed to handle message: %v", err)
			}
		}()
	}
}

func (c *TCPClient) handleMessage(b []byte) error {
	msg, err := messages.DeserializeMessage(b)
	if err != nil {
		return fmt.Errorf("failed to deserialize message: %v", err)
	}

	switch msg.Type {
	case messages.MessageTypeServerLoginSuccess:
		assignID := &messages.ServerLoginSuccess{}
		err := json.Unmarshal(msg.Payload, assignID)
		if err != nil {
			return fmt.Errorf("failed to deserialize server login success message: %v", err)
		}
		// write the client ID back to the manager
		c.clientIDChan <- assignID.ClientID
	case messages.MessageTypeServerLoginFailure:
		loginFailure := &messages.ServerLoginFailure{}
		err := json.Unmarshal(msg.Payload, loginFailure)
		if err != nil {
			return fmt.Errorf("failed to deserialize server login failure message: %v", err)
		}
		loginErr := fmt.Errorf("server login failure: %s", loginFailure.Reason)
		c.loginErrChan <- loginErr
	case messages.MessageTypeServerSyncTime:
		serverSyncTime := &messages.ServerSyncTime{}
		err := json.Unmarshal(msg.Payload, serverSyncTime)
		if err != nil {
			return fmt.Errorf("failed to deserialize server sync time message: %v", err)
		}
		c.serverTimeChan <- serverSyncTime
	case messages.MessageTypeServerPlayerConnect,
		messages.MessageTypeServerPlayerDisconnect,
		messages.MessageTypeServerNPCHit,
		messages.MessageTypeServerNPCKill,
		messages.MessageTypeServerPlayerHit,
		messages.MessageTypeServerPlayerKill:
		if err := c.messageQueue.Enqueue(msg); err != nil {
			return fmt.Errorf("failed to enqueue message: %v", err)
		}
	default:
		return fmt.Errorf("received unexpected message type from TCP server: %s", msg.Type)
	}

	return nil
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

// ReadFromTCP reads a buffer from a TCP connection
func ReadFromTCP(conn net.Conn) ([]byte, error) {
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

	return buf[:n], nil
}
