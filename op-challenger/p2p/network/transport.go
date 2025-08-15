package network

import (
	"fmt"
	"net"
	"time"
)

// Transport represents the network transport layer
type Transport interface {
	Start(address string) error
	Stop() error
	Accept() (Connection, error)
	Connect(address string) (Connection, error)
}

// Connection represents a network connection
type Connection interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Close() error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// TransportConfig contains configuration for transport layer
type TransportConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	BufferSize   int
}

// DefaultTransportConfig returns default transport configuration
func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		BufferSize:   4096,
	}
}

// TCPTransport implements Transport interface using TCP
type TCPTransport struct {
	listener    net.Listener
	config      TransportConfig
	connections map[string]*TCPConnection
}

// NewTransport creates a new transport layer
func NewTransport(config TransportConfig) (Transport, error) {
	return &TCPTransport{
		config:      config,
		connections: make(map[string]*TCPConnection),
	}, nil
}

// Start starts the TCP transport
func (t *TCPTransport) Start(address string) error {
	fmt.Printf("Starting TCP transport on %s\n", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to start TCP listener on %s: %v\n", address, err)
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	t.listener = listener
	fmt.Printf("TCP transport started successfully on %s\n", address)
	return nil
}

// Stop stops the TCP transport
func (t *TCPTransport) Stop() error {
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

// Accept accepts a new connection
func (t *TCPTransport) Accept() (Connection, error) {
	if t.listener == nil {
		return nil, fmt.Errorf("transport not started")
	}

	conn, err := t.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}

	// Set connection timeouts
	if err := conn.SetDeadline(time.Now().Add(t.config.ReadTimeout)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	tcpConn := &TCPConnection{
		conn:   conn,
		config: t.config,
	}

	return tcpConn, nil
}

// Connect connects to a remote address
func (t *TCPTransport) Connect(address string) (Connection, error) {
	fmt.Printf("Attempting to connect to %s\n", address)
	conn, err := net.DialTimeout("tcp", address, t.config.ReadTimeout)
	if err != nil {
		fmt.Printf("Failed to connect to %s: %v\n", address, err)
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	fmt.Printf("Successfully connected to %s\n", address)

	// Set connection timeouts
	if err := conn.SetDeadline(time.Now().Add(t.config.ReadTimeout)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	tcpConn := &TCPConnection{
		conn:   conn,
		config: t.config,
	}

	return tcpConn, nil
}

// TCPConnection implements Connection interface
type TCPConnection struct {
	conn   net.Conn
	config TransportConfig
}

// Read reads data from the connection
func (c *TCPConnection) Read() ([]byte, error) {
	buffer := make([]byte, c.config.BufferSize)

	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read from connection: %w", err)
	}

	return buffer[:n], nil
}

// Write writes data to the connection
func (c *TCPConnection) Write(data []byte) error {
	// Set write deadline
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to connection: %w", err)
	}

	return nil
}

// Close closes the connection
func (c *TCPConnection) Close() error {
	return c.conn.Close()
}

// RemoteAddr returns the remote network address
func (c *TCPConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr returns the local network address
func (c *TCPConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// SetDeadline sets the read and write deadlines
func (c *TCPConnection) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline
func (c *TCPConnection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (c *TCPConnection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
