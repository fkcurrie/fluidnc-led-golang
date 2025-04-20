package discovery

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/fkcurrie/fluidnc-led-golang/internal/types"
)

// Scanner represents a network scanner for discovering FluidNC devices
type Scanner struct {
	config types.DiscoveryConfig
}

// NewScanner creates a new network scanner
func NewScanner(config types.DiscoveryConfig) *Scanner {
	return &Scanner{
		config: config,
	}
}

// ScanResult represents the result of a network scan
type ScanResult struct {
	IPAddress string
	Port      int
	Valid     bool
	Error     error
}

// ScanNetwork scans the network for FluidNC devices
func (s *Scanner) ScanNetwork(ctx context.Context) ([]ScanResult, error) {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var results []ScanResult

	// Scan each interface
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get addresses for the interface
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Scan each address
		for _, addr := range addresses {
			// Skip non-IPv4 addresses
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}

			// Skip self-assigned addresses
			if ipNet.IP.IsLoopback() || ipNet.IP.IsLinkLocalUnicast() {
				continue
			}

			// Scan the network
			networkResults, err := s.scanNetworkRange(ctx, ipNet)
			if err != nil {
				continue
			}

			results = append(results, networkResults...)
		}
	}

	return results, nil
}

// scanNetworkRange scans a network range for FluidNC devices
func (s *Scanner) scanNetworkRange(ctx context.Context, ipNet *net.IPNet) ([]ScanResult, error) {
	var results []ScanResult

	// Get the network and broadcast addresses
	network := ipNet.IP.Mask(ipNet.Mask)
	broadcast := net.IP(make([]byte, 4))
	for i := range broadcast {
		broadcast[i] = network[i] | ^ipNet.Mask[i]
	}

	// Create a channel for results
	resultChan := make(chan ScanResult, 256)

	// Start scanning
	for i := 1; i < 255; i++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		// Create IP address
		ip := make(net.IP, 4)
		copy(ip, network)
		ip[3] = byte(i)

		// Skip network and broadcast addresses
		if ip.Equal(network) || ip.Equal(broadcast) {
			continue
		}

		// Start a goroutine to scan this IP
		go s.scanIP(ctx, ip, resultChan)
	}

	// Collect results
	timeout := time.After(time.Duration(s.config.Timeout) * time.Second)
	for i := 0; i < 254; i++ {
		select {
		case result := <-resultChan:
			if result.Valid {
				results = append(results, result)
			}
		case <-timeout:
			return results, nil
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}

	return results, nil
}

// scanIP scans a single IP address for FluidNC devices
func (s *Scanner) scanIP(ctx context.Context, ip net.IP, resultChan chan<- ScanResult) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.config.Timeout)*time.Second)
	defer cancel()

	// Try to connect to the WebSocket port
	address := net.JoinHostPort(ip.String(), strconv.Itoa(81))
	conn, err := net.DialTimeout("tcp", address, time.Duration(s.config.Timeout)*time.Second)
	if err != nil {
		resultChan <- ScanResult{
			IPAddress: ip.String(),
			Port:      81,
			Valid:     false,
			Error:     err,
		}
		return
	}
	defer conn.Close()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	// Try to validate if this is a FluidNC device
	// This is a simple check - we could make it more sophisticated
	valid := s.validateFluidNC(conn)

	resultChan <- ScanResult{
		IPAddress: ip.String(),
		Port:      81,
		Valid:     valid,
		Error:     nil,
	}
}

// validateFluidNC validates if a connection is to a FluidNC device
func (s *Scanner) validateFluidNC(conn net.Conn) bool {
	// This is a simple validation - we could make it more sophisticated
	// For now, we just check if the port is open
	return true
} 