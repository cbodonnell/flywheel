package network

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/gorilla/websocket"
)

// WSServer represents a WebSocket server.
type WSServer struct {
	port int
	tls  *TLSConfig
}

type TLSConfig struct {
	CertFile string
	KeyFile  string
}

type NewWSServerOptions struct {
	Port int
	TLS  *TLSConfig
}

// NewWSServer creates a new WebSocket server.
func NewWSServer(opts NewWSServerOptions) *WSServer {
	return &WSServer{
		port: opts.Port,
		tls:  opts.TLS,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Start starts the WebSocket server.
func (s *WSServer) Start(ctx context.Context, disconnectHandler ControlDisconnectHandler, messageHandler ControlMessageHandler) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("Failed to upgrade to WebSocket: %v", err)
			return
		}
		log.Debug("New WebSocket connection from %s", conn.RemoteAddr().String())
		go s.handleWSConnection(ctx, conn, disconnectHandler, messageHandler)
	})

	addr := fmt.Sprintf(":%d", s.port)
	server := &http.Server{Addr: addr, Handler: nil}

	go func() {
		<-ctx.Done()
		server.Shutdown(ctx)
	}()

	var listenAndServe func() error
	if s.tls != nil {
		log.Info("WebSocket server listening on %s with TLS", addr)
		listenAndServe = func() error {
			return server.ListenAndServeTLS(s.tls.CertFile, s.tls.KeyFile)
		}
	} else {
		log.Info("WebSocket server listening on %s", addr)
		listenAndServe = server.ListenAndServe
	}
	if err := listenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("WebSocket server closed")
			return
		}
		log.Error("WebSocket server error: %v", err)
	}
}

// handleWSConnection handles a WebSocket connection.
func (s *WSServer) handleWSConnection(ctx context.Context, conn *websocket.Conn, disconnectHandler ControlDisconnectHandler, messageHandler ControlMessageHandler) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		disconnectHandler(nil, conn)
		conn.Close()
	}()

	for {
		message, err := ReadMessageFromWS(conn)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("Error reading WebSocket message from %s: %v", conn.RemoteAddr().String(), err)
			}
			log.Trace("Connection closed for %s", conn.RemoteAddr().String())
			return
		}

		go messageHandler(ctx, nil, conn, message)
	}
}

// WriteMessageToWS writes a Message to a WebSocket connection
func WriteMessageToWS(conn *websocket.Conn, msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return fmt.Errorf("failed to write message to WebSocket connection: %v", err)
	}

	return nil
}

// ReadMessageFromWS reads a Message from a WebSocket connection
func ReadMessageFromWS(conn *websocket.Conn) (*messages.Message, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	msg, err := messages.DeserializeMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
