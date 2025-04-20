package rpi5matrix

import (
	"fmt"
	"image"
	"image/color"
	"sync"
)

// Matrix represents an RGB LED matrix display
type Matrix struct {
	width      int
	height     int
	brightness int
	gpioPin    int
	strip      *RGBMatrix
	mu         sync.RWMutex
}

// Config holds the configuration for the LED matrix
type Config struct {
	Width      int
	Height     int
	Brightness int
	GPIOPin    int
}

// NewMatrix creates a new LED matrix display
func NewMatrix(cfg *Config) (*Matrix, error) {
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", cfg.Width, cfg.Height)
	}

	if cfg.Brightness < 0 || cfg.Brightness > 255 {
		return nil, fmt.Errorf("brightness must be between 0 and 255")
	}

	// Create the RGB matrix
	strip, err := NewRGBMatrix(cfg.GPIOPin, cfg.Width, cfg.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to create RGB matrix: %v", err)
	}

	// Set initial brightness
	if err := strip.SetBrightness(cfg.Brightness); err != nil {
		strip.Close()
		return nil, fmt.Errorf("failed to set brightness: %v", err)
	}

	return &Matrix{
		width:      cfg.Width,
		height:     cfg.Height,
		brightness: cfg.Brightness,
		gpioPin:    cfg.GPIOPin,
		strip:      strip,
	}, nil
}

// Close closes the LED matrix
func (m *Matrix) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.strip != nil {
		return m.strip.Close()
	}
	return nil
}

// Clear clears all LEDs
func (m *Matrix) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.Clear()
}

// SetPixel sets a pixel at the given coordinates to the given color
func (m *Matrix) SetPixel(x, y int, c color.Color) error {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return fmt.Errorf("coordinates out of bounds: (%d, %d)", x, y)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate the LED index based on the serpentine pattern
	var index int
	if y%2 == 0 {
		index = y*m.width + x
	} else {
		index = y*m.width + (m.width - 1 - x)
	}

	return m.strip.SetPixel(index, c)
}

// Show updates the display with the current buffer
func (m *Matrix) Show() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.Show()
}

// SetBrightness sets the brightness of the LED matrix
func (m *Matrix) SetBrightness(brightness int) error {
	if brightness < 0 || brightness > 255 {
		return fmt.Errorf("brightness must be between 0 and 255")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.strip.SetBrightness(brightness); err != nil {
		return fmt.Errorf("failed to set brightness: %v", err)
	}

	m.brightness = brightness
	return nil
}

// GetBrightness returns the current brightness of the LED matrix
func (m *Matrix) GetBrightness() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.brightness
}

// GetDimensions returns the dimensions of the LED matrix
func (m *Matrix) GetDimensions() (width, height int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.width, m.height
}

// Fill fills the entire matrix with a color
func (m *Matrix) Fill(c color.Color) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.Fill(c)
}

// Scroll scrolls the display by the given number of pixels
func (m *Matrix) Scroll(dx, dy int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.Scroll(dx, dy)
}

// SetImage sets the display to show an image
func (m *Matrix) SetImage(img image.Image) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.SetImage(img)
}

// SetText sets the display to show text
func (m *Matrix) SetText(text string, x, y int, c color.Color) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.SetText(text, x, y, c)
}

// SetFont sets the font for text rendering
func (m *Matrix) SetFont(font interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.SetFont(font)
}

// SetRotation sets the rotation of the display
func (m *Matrix) SetRotation(rotation int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.strip.SetRotation(rotation)
}

// GetRotation returns the current rotation of the display
func (m *Matrix) GetRotation() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.strip.GetRotation()
}

// SetPixelColor sets a pixel at the given coordinates to the given color
func (m *Matrix) SetPixelColor(x, y int, r, g, b uint8) error {
	return m.SetPixel(x, y, color.RGBA{r, g, b, 255})
}

// GetPixelColor gets the color of a pixel at the given coordinates
func (m *Matrix) GetPixelColor(x, y int) (r, g, b uint8, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return 0, 0, 0, fmt.Errorf("coordinates out of bounds: (%d, %d)", x, y)
	}

	var index int
	if y%2 == 0 {
		index = y*m.width + x
	} else {
		index = y*m.width + (m.width - 1 - x)
	}

	return m.strip.GetPixelColor(index)
}

// SetPixelBrightness sets the brightness of a single pixel
func (m *Matrix) SetPixelBrightness(x, y int, brightness uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return fmt.Errorf("coordinates out of bounds: (%d, %d)", x, y)
	}

	var index int
	if y%2 == 0 {
		index = y*m.width + x
	} else {
		index = y*m.width + (m.width - 1 - x)
	}

	return m.strip.SetPixelBrightness(index, brightness)
}

// GetPixelBrightness gets the brightness of a single pixel
func (m *Matrix) GetPixelBrightness(x, y int) (uint8, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return 0, fmt.Errorf("coordinates out of bounds: (%d, %d)", x, y)
	}

	var index int
	if y%2 == 0 {
		index = y*m.width + x
	} else {
		index = y*m.width + (m.width - 1 - x)
	}

	return m.strip.GetPixelBrightness(index)
}

// SetPixelHSV sets a pixel at the given coordinates using HSV color values
func (m *Matrix) SetPixelHSV(x, y int, h, s, v float64) error {
	return m.SetPixel(x, y, hsvToRGB(h, s, v))
}

// hsvToRGB converts HSV color values to RGB
func hsvToRGB(h, s, v float64) color.Color {
	// This is a placeholder - in a real implementation, this would convert
	// HSV to RGB
	return color.RGBA{0, 0, 0, 255}
} 