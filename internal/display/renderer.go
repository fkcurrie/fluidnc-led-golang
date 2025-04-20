package display

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"sync"
	"time"

	"github.com/fkcurrie/fluidnc-led-golang/internal/config"
	"github.com/fkcurrie/fluidnc-led-golang/internal/types"
)

// Renderer handles the display rendering logic
type Renderer struct {
	cfg    *config.DisplayConfig
	matrix types.Matrix
	mu     sync.RWMutex
}

// NewRenderer creates a new renderer instance
func NewRenderer(cfg *config.DisplayConfig) *Renderer {
	return &Renderer{
		cfg: cfg,
	}
}

// SetMatrix sets the matrix to render to
func (r *Renderer) SetMatrix(matrix types.Matrix) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matrix = matrix
}

// Start starts the renderer
func (r *Renderer) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(r.cfg.RefreshRate) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.render(); err != nil {
				log.Printf("Failed to render: %v", err)
			}
		}
	}
}

// render renders the current state to the matrix
func (r *Renderer) render() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.matrix == nil {
		return nil
	}

	// TODO: Implement actual rendering logic
	// For now, just clear the matrix
	return r.matrix.Clear()
}

// GetDisplayLayout returns the layout for the display
func (r *Renderer) GetDisplayLayout(data types.DisplayData) DisplayLayout {
	return DisplayLayout{
		IPAddress: IPAddressLayout{
			X: r.cfg.Width - 10,
			Y: 0,
			Color: color.RGBA{
				R: 255,
				G: 255,
				B: 255,
				A: 255,
			},
		},
		Coordinates: CoordinatesLayout{
			X: XCoordinateLayout{
				X: 0,
				Y: 5,
				Color: color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 255,
				},
			},
			Y: YCoordinateLayout{
				X: 0,
				Y: 15,
				Color: color.RGBA{
					R: 0,
					G: 255,
					B: 0,
					A: 255,
				},
			},
			Z: ZCoordinateLayout{
				X: 0,
				Y: 25,
				Color: color.RGBA{
					R: 0,
					G: 0,
					B: 255,
					A: 255,
				},
			},
		},
		Status: StatusLayout{
			X: 20,
			Y: 25,
			Color: color.RGBA{
				R: 255,
				G: 255,
				B: 255,
				A: 255,
			},
		},
		ConnectionIndicator: ConnectionIndicatorLayout{
			X: r.cfg.Width - 2,
			Y: 0,
			Connected: data.Connected,
			Color: color.RGBA{
				R: 0,
				G: 255,
				B: 0,
				A: 255,
			},
		},
	}
}

// DisplayLayout represents the layout for the display
type DisplayLayout struct {
	IPAddress           IPAddressLayout
	Coordinates         CoordinatesLayout
	Status              StatusLayout
	ConnectionIndicator ConnectionIndicatorLayout
}

// IPAddressLayout represents the layout for the IP address
type IPAddressLayout struct {
	X     int
	Y     int
	Color color.Color
}

// CoordinatesLayout represents the layout for the coordinates
type CoordinatesLayout struct {
	X XCoordinateLayout
	Y YCoordinateLayout
	Z ZCoordinateLayout
}

// XCoordinateLayout represents the layout for the X coordinate
type XCoordinateLayout struct {
	X     int
	Y     int
	Color color.Color
}

// YCoordinateLayout represents the layout for the Y coordinate
type YCoordinateLayout struct {
	X     int
	Y     int
	Color color.Color
}

// ZCoordinateLayout represents the layout for the Z coordinate
type ZCoordinateLayout struct {
	X     int
	Y     int
	Color color.Color
}

// StatusLayout represents the layout for the status
type StatusLayout struct {
	X     int
	Y     int
	Color color.Color
}

// ConnectionIndicatorLayout represents the layout for the connection indicator
type ConnectionIndicatorLayout struct {
	X         int
	Y         int
	Connected bool
	Color     color.Color
} 