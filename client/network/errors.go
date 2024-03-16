package network

// ErrConnectionClosedByServer is returned when the TCP connection is closed
type ErrConnectionClosedByServer struct{}

func (e *ErrConnectionClosedByServer) Error() string {
	return "connection closed by server"
}

// ErrConnectionClosed is returned when the TCP connection is closed
type ErrConnectionClosedByClient struct{}

func (e *ErrConnectionClosedByClient) Error() string {
	return "connection closed by client"
}
