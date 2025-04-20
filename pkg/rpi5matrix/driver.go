package rpi5matrix

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"
	"time"

	"github.com/fcurrie/fluidnc-led-golang/pkg/gpio"
	"github.com/fcurrie/fluidnc-led-golang/pkg/mmap"
	"github.com/fcurrie/fluidnc-led-golang/pkg/pio"
)

const (
	// Default configuration for RGB Matrix
	DefaultWidth  = 32
	DefaultHeight = 8
	DefaultPin    = 18 // GPIO18 is used by default on the Bonnet/HAT
	// PIO base address for Raspberry Pi 5
	PIOBaseAddr = 0x50200000
	// PIO size in bytes
	PIOSize = 0x1000
	// Number of PIO state machines
	NumStateMachines = 4
	// HUB75 protocol timing (in nanoseconds)
	HUB75Timing = 100
)

// RGBMatrix represents an RGB LED matrix display
type RGBMatrix struct {
	width      int
	height     int
	brightness int
	pin        *gpio.Pin
	pio        *pio.PIOState
	mem        *mmap.MemoryMap
	mutex      sync.Mutex
	buffer     []color.Color
}

// NewRGBMatrix creates a new RGB matrix display
func NewRGBMatrix(width, height int, pin int) (*RGBMatrix, error) {
	// Create GPIO pin
	gpioPin, err := gpio.NewPin(pin)
	if err != nil {
		return nil, fmt.Errorf("failed to create GPIO pin: %v", err)
	}

	// Map PIO memory
	mem, err := mmap.NewMemoryMap(PIOBaseAddr, PIOSize)
	if err != nil {
		gpioPin.Close()
		return nil, fmt.Errorf("failed to map PIO memory: %v", err)
	}

	// Create PIO state machine
	pioState, err := pio.NewPIOState(mem, 0) // Use first state machine
	if err != nil {
		mem.Close()
		gpioPin.Close()
		return nil, fmt.Errorf("failed to create PIO state: %v", err)
	}

	// Initialize buffer
	buffer := make([]color.Color, width*height)

	return &RGBMatrix{
		width:  width,
		height: height,
		pin:    gpioPin,
		pio:    pioState,
		mem:    mem,
		buffer: buffer,
	}, nil
}

// Close closes the RGB matrix display
func (m *RGBMatrix) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.pio.Stop(); err != nil {
		return fmt.Errorf("failed to stop PIO: %v", err)
	}

	if err := m.mem.Close(); err != nil {
		return fmt.Errorf("failed to close memory map: %v", err)
	}

	if err := m.pin.Close(); err != nil {
		return fmt.Errorf("failed to close GPIO pin: %v", err)
	}

	return nil
}

// SetBrightness sets the brightness of the LED matrix
func (m *RGBMatrix) SetBrightness(brightness int) error {
	if brightness < 0 || brightness > 255 {
		return fmt.Errorf("brightness must be between 0 and 255")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.brightness = brightness
	return nil
}

// GetBrightness returns the current brightness
func (m *RGBMatrix) GetBrightness() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.brightness
}

// Clear clears the display
func (m *RGBMatrix) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.buffer {
		m.buffer[i] = color.Black
	}

	return m.show()
}

// SetPixel sets a pixel's color
func (m *RGBMatrix) SetPixel(x, y int, c color.Color) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return fmt.Errorf("pixel coordinates out of bounds")
	}

	index := y*m.width + x
	m.buffer[index] = c

	return nil
}

// GetPixelColor gets the color of a pixel at the given index
func (m *RGBMatrix) GetPixelColor(index int) (uint8, uint8, uint8, error) {
	if index < 0 || index >= len(m.buffer) {
		return 0, 0, 0, fmt.Errorf("index out of bounds: %d", index)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	r, g, b, _ := m.buffer[index].RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), nil
}

// Show updates the display with the current buffer contents
func (m *RGBMatrix) Show() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.show()
}

// show is an internal method that assumes the mutex is already locked
func (m *RGBMatrix) show() error {
	// Convert buffer to HUB75 protocol data
	data := make([]byte, len(m.buffer)*3)
	for i, c := range m.buffer {
		r, g, b, _ := c.RGBA()
		offset := i * 3
		data[offset] = byte(r >> 8)
		data[offset+1] = byte(g >> 8)
		data[offset+2] = byte(b >> 8)
	}

	// Write data to PIO FIFO
	if err := m.pio.WriteFIFO(data); err != nil {
		return fmt.Errorf("failed to write to PIO FIFO: %v", err)
	}

	// Start PIO state machine
	if err := m.pio.Start(); err != nil {
		return fmt.Errorf("failed to start PIO: %v", err)
	}

	// Wait for data to be processed
	time.Sleep(time.Duration(HUB75Timing) * time.Nanosecond)

	// Stop PIO state machine
	if err := m.pio.Stop(); err != nil {
		return fmt.Errorf("failed to stop PIO: %v", err)
	}

	return nil
}

// Fill fills the entire matrix with a color
func (m *RGBMatrix) Fill(c color.Color) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.buffer {
		m.buffer[i] = c
	}

	return m.show()
}

// Scroll scrolls the display by the given number of pixels
func (m *RGBMatrix) Scroll(dx, dy int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a new buffer for the scrolled content
	newBuffer := make([]color.Color, len(m.buffer))

	// Copy the content with offset
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			srcX := (x + dx + m.width) % m.width
			srcY := (y + dy + m.height) % m.height
			srcIndex := srcY*m.width + srcX
			dstIndex := y*m.width + x
			newBuffer[dstIndex] = m.buffer[srcIndex]
		}
	}

	m.buffer = newBuffer
	return m.show()
}

// SetImage sets the display to show an image
func (m *RGBMatrix) SetImage(img image.Image) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bounds := img.Bounds()
	if bounds.Dx() != m.width || bounds.Dy() != m.height {
		return fmt.Errorf("image dimensions (%dx%d) do not match matrix dimensions (%dx%d)",
			bounds.Dx(), bounds.Dy(), m.width, m.height)
	}

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			m.buffer[y*m.width+x] = img.At(x, y)
		}
	}

	return m.show()
}

// SetText sets the display to show text
func (m *RGBMatrix) SetText(text string, x, y int, c color.Color) error {
	// This is a placeholder - in a real implementation, this would render text
	// using a font and set the pixels accordingly
	return fmt.Errorf("SetText not implemented")
}

// SetFont sets the font for text rendering
func (m *RGBMatrix) SetFont(font interface{}) error {
	// This is a placeholder - in a real implementation, this would set the font
	// for text rendering
	return fmt.Errorf("SetFont not implemented")
}

// SetRotation sets the rotation of the display
func (m *RGBMatrix) SetRotation(rotation int) error {
	// This is a placeholder - in a real implementation, this would set the
	// rotation of the display
	return fmt.Errorf("SetRotation not implemented")
}

// GetRotation returns the current rotation of the display
func (m *RGBMatrix) GetRotation() int {
	// This is a placeholder - in a real implementation, this would return the
	// current rotation of the display
	return 0
}

// SetPixelBrightness sets the brightness of a single pixel
func (m *RGBMatrix) SetPixelBrightness(index int, brightness uint8) error {
	if index < 0 || index >= len(m.buffer) {
		return fmt.Errorf("index out of bounds: %d", index)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	r, g, b, _ := m.buffer[index].RGBA()
	r = uint32(brightness) * r / 255
	g = uint32(brightness) * g / 255
	b = uint32(brightness) * b / 255

	m.buffer[index] = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
	return nil
}

// GetPixelBrightness gets the brightness of a single pixel
func (m *RGBMatrix) GetPixelBrightness(index int) (uint8, error) {
	if index < 0 || index >= len(m.buffer) {
		return 0, fmt.Errorf("index out of bounds: %d", index)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	r, g, b, _ := m.buffer[index].RGBA()
	brightness := (uint32(r) + uint32(g) + uint32(b)) / 3
	return uint8(brightness >> 8), nil
} 