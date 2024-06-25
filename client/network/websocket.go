package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"nhooyr.io/websocket"
)

// WSClient represents a WebSocket client.
type WSClient struct {
	serverAddr     string
	messageQueue   queue.Queue
	clientIDChan   chan<- uint32
	loginErrChan   chan<- error
	serverTimeChan chan<- *messages.ServerSyncTime
	conn           *websocket.Conn
}

// NewWSClient creates a new WebSocket client.
func NewWSClient(serverAddr string, messageQueue queue.Queue, clientIDChan chan<- uint32, loginErrChan chan<- error, serverTimeChan chan<- *messages.ServerSyncTime) *WSClient {
	return &WSClient{
		serverAddr:     serverAddr,
		messageQueue:   messageQueue,
		clientIDChan:   clientIDChan,
		loginErrChan:   loginErrChan,
		serverTimeChan: serverTimeChan,
	}
}

// Connect establishes a connection to the WebSocket server.
func (c *WSClient) Connect(ctx context.Context) error {
	log.Info("Connecting to WebSocket server at %s", c.serverAddr)
	conn, _, err := websocket.Dial(ctx, c.serverAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	return nil
}

// HandleMessages handles incoming messages from the WebSocket server.
func (c *WSClient) HandleMessages(ctx context.Context) error {
	defer c.conn.Close(websocket.StatusNormalClosure, "client closed connection")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, message, err := c.conn.Read(ctx)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					log.Info("WebSocket server closed")
					return nil
				} else if strings.Contains(err.Error(), "context canceled") {
					log.Info("WebSocket connection closed by client")
					return nil
				}
				log.Error("Failed to read message from WebSocket connection: %v", err)
				return err
			}

			go func() {
				if err := c.handleMessage(message); err != nil {
					log.Error("Failed to handle message: %v", err)
				}
			}()
		}
	}
}

// handleMessage processes a received message.
func (c *WSClient) handleMessage(b []byte) error {
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
		messages.MessageTypeServerPlayerKill,
		messages.MessageTypeServerGameUpdate,
		messages.MessageTypeServerPlayerUpdate,
		messages.MessageTypeServerNPCUpdate:
		if err := c.messageQueue.Enqueue(msg); err != nil {
			return fmt.Errorf("failed to enqueue message: %v", err)
		}
	case messages.MessageTypeServerPong:
		log.Debug("Received server pong")
	default:
		return fmt.Errorf("received unexpected message type from WebSocket server: %s", msg.Type)
	}

	return nil
}

// Close closes the WebSocket connection.
func (c *WSClient) Close() error {
	if c.conn == nil {
		log.Warn("WebSocket connection is already closed")
		return nil
	}
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

// SendMessage sends a message to the WebSocket server.
func (c *WSClient) SendMessage(msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	// TODO: can we use a context here?
	if err := c.conn.Write(context.TODO(), websocket.MessageBinary, b); err != nil {
		return fmt.Errorf("failed to write message to WebSocket connection: %v", err)
	}

	return nil
}
