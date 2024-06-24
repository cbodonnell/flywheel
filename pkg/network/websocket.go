package network

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"nhooyr.io/websocket"
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

// Start starts the WebSocket server.
func (s *WSServer) Start(ctx context.Context, messageChan chan<- *messages.Message, loginChan chan<- *LoginEvent, logoutChan chan<- *LogoutEvent) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"}, // TODO: restrict origins all around once this is working
		})
		if err != nil {
			log.Error("Failed to accept WebSocket connection: %v", err)
			return
		}

		go s.handleWSConnection(ctx, conn, messageChan, loginChan, logoutChan)
	})

	addr := fmt.Sprintf(":%d", s.port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

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
func (s *WSServer) handleWSConnection(ctx context.Context, conn *websocket.Conn, messageChan chan<- *messages.Message, loginChan chan<- *LoginEvent, logoutChan chan<- *LogoutEvent) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		logoutChan <- &LogoutEvent{
			WSConn: conn,
		}
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		message, err := ReadMessageFromWS(ctx, conn)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Trace("WebSocket connection closed by client")
				return
			}

			log.Error("Failed to read message from WebSocket connection: %v", err)
			return
		}

		if message.ClientID != 0 {
			messageChan <- message
			continue
		}

		if message.Type == messages.MessageTypeClientLogin {
			loginChan <- &LoginEvent{
				WSConn:  conn,
				Message: message,
			}
			continue
		}

		log.Warn("Received a %s message from an unknown client", message.Type)
	}
}

// WriteMessageToWS writes a Message to a WebSocket connection
func WriteMessageToWS(ctx context.Context, conn *websocket.Conn, msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	if err := conn.Write(ctx, websocket.MessageBinary, b); err != nil {
		return fmt.Errorf("failed to write message to WebSocket connection: %v", err)
	}

	return nil
}

// ReadMessageFromWS reads a Message from a WebSocket connection
func ReadMessageFromWS(ctx context.Context, conn *websocket.Conn) (*messages.Message, error) {
	_, message, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := messages.DeserializeMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
