package fluidnc

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fkcurrie/fluidnc-led-golang/internal/types"
	"github.com/gorilla/websocket"
)

// Client represents a FluidNC WebSocket client
type Client struct {
	config     types.FluidNCConfig
	conn       *websocket.Conn
	statusChan chan types.MachineStatus
	done       chan struct{}
}

// NewClient creates a new FluidNC WebSocket client
func NewClient(config types.FluidNCConfig) *Client {
	return &Client{
		config:     config,
		statusChan: make(chan types.MachineStatus, 10),
		done:       make(chan struct{}),
	}
}

// Connect connects to the FluidNC WebSocket server
func (c *Client) Connect(ctx context.Context) error {
	// Create WebSocket URL
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Path:   "/",
	}

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to FluidNC: %w", err)
	}

	c.conn = conn

	// Start goroutines for reading and writing
	go c.readPump(ctx)
	go c.writePump(ctx)

	return nil
}

// Disconnect disconnects from the FluidNC WebSocket server
func (c *Client) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Status returns a channel that receives machine status updates
func (c *Client) Status() <-chan types.MachineStatus {
	return c.statusChan
}

// Close closes the client
func (c *Client) Close() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}

// readPump pumps messages from the WebSocket connection to the status channel
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}

			// Parse the message
			status, err := parseStatusMessage(string(message))
			if err != nil {
				log.Printf("error parsing status message: %v", err)
				continue
			}

			// Send the status to the channel
			select {
			case c.statusChan <- status:
			default:
				// Channel is full, skip this update
			}
		}
	}
}

// writePump pumps messages from the status channel to the WebSocket connection
func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		default:
			// Send status request
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte("?")); err != nil {
				return
			}

			// Wait for the status interval
			time.Sleep(time.Duration(c.config.StatusInterval*1000) * time.Millisecond)
		}
	}
}

// parseStatusMessage parses a status message from FluidNC
func parseStatusMessage(message string) (types.MachineStatus, error) {
	// Example message: <Idle|MPos:0.000,0.000,0.000|Bf:15,100|F:0|FS:0,0>
	status := types.MachineStatus{
		LastUpdated: time.Now(),
	}

	// Remove < and >
	message = strings.TrimPrefix(message, "<")
	message = strings.TrimSuffix(message, ">")

	// Split by |
	parts := strings.Split(message, "|")
	if len(parts) < 1 {
		return status, fmt.Errorf("invalid message format")
	}

	// Parse state
	status.State = types.MachineState(parts[0])

	// Parse coordinates
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if strings.HasPrefix(part, "MPos:") {
			coords := strings.Split(strings.TrimPrefix(part, "MPos:"), ",")
			if len(coords) >= 3 {
				status.Coordinates.X = parseFloat(coords[0])
				status.Coordinates.Y = parseFloat(coords[1])
				status.Coordinates.Z = parseFloat(coords[2])
			}
		} else if strings.HasPrefix(part, "F:") {
			status.FeedRate = parseFloat(strings.TrimPrefix(part, "F:"))
		} else if strings.HasPrefix(part, "S:") {
			status.SpindleSpeed = parseFloat(strings.TrimPrefix(part, "S:"))
		} else if strings.HasPrefix(part, "Bf:") {
			buf := strings.Split(strings.TrimPrefix(part, "Bf:"), ",")
			if len(buf) >= 2 {
				status.BufferState = parseInt(buf[0])
			}
		} else if strings.HasPrefix(part, "Ln:") {
			status.LineNumber = parseInt(strings.TrimPrefix(part, "Ln:"))
		}
	}

	return status, nil
}

// parseFloat parses a float from a string
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseInt parses an int from a string
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
} 