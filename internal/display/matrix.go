package display

import (
	"fmt"
	"time"

	"github.com/fluidnc-led/internal/config"
	"github.com/rpi-ws281x/rpi-ws281x-go"
)

// LEDMatrix represents an LED matrix display
type LEDMatrix struct {
	config *config.DisplayConfig
	strip  *ws2811.WS2811
}

// NewLEDMatrix creates a new LED matrix display
func NewLEDMatrix(cfg *config.DisplayConfig) (*LEDMatrix, error) {
	// Create WS2811 configuration
	ws2811Config := ws2811.DefaultConfig
	ws2811Config.Channels[0].Brightness = cfg.Brightness
	ws2811Config.Channels[0].GpioPin = cfg.GPIOPin
	ws2811Config.Channels[0].LedCount = cfg.Width * cfg.Height
	ws2811Config.Channels[0].StripeType = ws2811.WS2811StripGRB

	// Initialize WS2811
	strip, err := ws2811.MakeWS2811(&ws2811Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create WS2811: %v", err)
	}

	// Initialize the strip
	if err := strip.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize WS2811: %v", err)
	}

	return &LEDMatrix{
		config: cfg,
		strip:  strip,
	}, nil
}

// Close closes the LED matrix
func (m *LEDMatrix) Close() error {
	if m.strip != nil {
		m.strip.Fini()
	}
	return nil
}

// Clear clears all LEDs
func (m *LEDMatrix) Clear() error {
	for i := 0; i < m.config.Width*m.config.Height; i++ {
		m.strip.Leds(0)[i] = 0
	}
	return m.strip.Render()
}

// SetPixel sets a pixel at the given coordinates to the given color
func (m *LEDMatrix) SetPixel(x, y int, color uint32) error {
	if x < 0 || x >= m.config.Width || y < 0 || y >= m.config.Height {
		return fmt.Errorf("coordinates out of bounds: (%d, %d)", x, y)
	}

	// Calculate the LED index based on the serpentine pattern
	var index int
	if y%2 == 0 {
		index = y*m.config.Width + x
	} else {
		index = y*m.config.Width + (m.config.Width - 1 - x)
	}

	m.strip.Leds(0)[index] = color
	return nil
}

// Render renders the current state of the LED matrix
func (m *LEDMatrix) Render() error {
	return m.strip.Render()
}

// SetBrightness sets the brightness of the LED matrix
func (m *LEDMatrix) SetBrightness(brightness int) error {
	if brightness < 0 || brightness > 255 {
		return fmt.Errorf("brightness must be between 0 and 255")
	}

	m.config.Brightness = brightness
	m.strip.SetBrightness(0, brightness)
	return m.strip.Render()
}

// GetBrightness returns the current brightness of the LED matrix
func (m *LEDMatrix) GetBrightness() int {
	return m.config.Brightness
}

// GetDimensions returns the dimensions of the LED matrix
func (m *LEDMatrix) GetDimensions() (width, height int) {
	return m.config.Width, m.config.Height
} 