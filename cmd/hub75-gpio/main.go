package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// Constants for display size
const (
	DISPLAY_WIDTH  = 64 // Width in pixels for a 32x64 panel
	DISPLAY_HEIGHT = 32 // Height in pixels for a 32x64 panel
	ROWS          = 16  // Number of addressable rows (DISPLAY_HEIGHT/2 for panels with upper/lower half data)
	FONT_HEIGHT   = 7   // Height of our font in pixels
	FONT_WIDTH    = 5   // Width of each character in our font
	CHAR_SPACING  = 1   // Space between characters
)

// Simple 5x7 font for scrolling text
// Each character is represented by 5 bytes, each byte representing a column of pixels
// Where 1 bits are "on" pixels
var font5x7 = map[rune][]byte{
	'A': {0x7E, 0x09, 0x09, 0x09, 0x7E},
	'B': {0x7F, 0x49, 0x49, 0x49, 0x36},
	'C': {0x3E, 0x41, 0x41, 0x41, 0x22},
	'D': {0x7F, 0x41, 0x41, 0x22, 0x1C},
	'E': {0x7F, 0x49, 0x49, 0x49, 0x41},
	'F': {0x7F, 0x09, 0x09, 0x09, 0x01},
	'G': {0x3E, 0x41, 0x49, 0x49, 0x3A},
	'H': {0x7F, 0x08, 0x08, 0x08, 0x7F},
	'I': {0x00, 0x41, 0x7F, 0x41, 0x00},
	'J': {0x20, 0x40, 0x41, 0x3F, 0x01},
	'K': {0x7F, 0x08, 0x14, 0x22, 0x41},
	'L': {0x7F, 0x40, 0x40, 0x40, 0x40},
	'M': {0x7F, 0x02, 0x0C, 0x02, 0x7F},
	'N': {0x7F, 0x04, 0x08, 0x10, 0x7F},
	'O': {0x3E, 0x41, 0x41, 0x41, 0x3E},
	'P': {0x7F, 0x09, 0x09, 0x09, 0x06},
	'Q': {0x3E, 0x41, 0x51, 0x21, 0x5E},
	'R': {0x7F, 0x09, 0x19, 0x29, 0x46},
	'S': {0x26, 0x49, 0x49, 0x49, 0x32},
	'T': {0x01, 0x01, 0x7F, 0x01, 0x01},
	'U': {0x3F, 0x40, 0x40, 0x40, 0x3F},
	'V': {0x1F, 0x20, 0x40, 0x20, 0x1F},
	'W': {0x3F, 0x40, 0x30, 0x40, 0x3F},
	'X': {0x63, 0x14, 0x08, 0x14, 0x63},
	'Y': {0x07, 0x08, 0x70, 0x08, 0x07},
	'Z': {0x61, 0x51, 0x49, 0x45, 0x43},
	'0': {0x3E, 0x51, 0x49, 0x45, 0x3E},
	'1': {0x00, 0x42, 0x7F, 0x40, 0x00},
	'2': {0x42, 0x61, 0x51, 0x49, 0x46},
	'3': {0x21, 0x41, 0x45, 0x4B, 0x31},
	'4': {0x18, 0x14, 0x12, 0x7F, 0x10},
	'5': {0x27, 0x45, 0x45, 0x45, 0x39},
	'6': {0x3C, 0x4A, 0x49, 0x49, 0x30},
	'7': {0x01, 0x71, 0x09, 0x05, 0x03},
	'8': {0x36, 0x49, 0x49, 0x49, 0x36},
	'9': {0x06, 0x49, 0x49, 0x29, 0x1E},
	' ': {0x00, 0x00, 0x00, 0x00, 0x00},
	'!': {0x00, 0x00, 0x5F, 0x00, 0x00},
	'.': {0x00, 0x60, 0x60, 0x00, 0x00},
	',': {0x00, 0x50, 0x30, 0x00, 0x00},
	':': {0x00, 0x36, 0x36, 0x00, 0x00},
	'-': {0x08, 0x08, 0x08, 0x08, 0x08},
	'+': {0x08, 0x08, 0x3E, 0x08, 0x08},
}

// HUB75 pin configuration for Adafruit RGB Matrix Bonnet
type HUB75Config struct {
	R1Pin  int // Red data for upper half
	G1Pin  int // Green data for upper half
	B1Pin  int // Blue data for upper half
	R2Pin  int // Red data for lower half
	G2Pin  int // Green data for lower half
	B2Pin  int // Blue data for lower half
	CLKPin int // Clock signal
	OEPin  int // Output enable
	LAPin  int // Latch signal
	ABPin  int // Address bit A
	BCPin  int // Address bit B
	CCPin  int // Address bit C
	DPin   int // Address bit D
	EPin   int // Address bit E
}

// HUB75Controller manages the pins for the HUB75 LED matrix
type HUB75Controller struct {
	config  HUB75Config
	lines   map[int]*gpiocdev.Line
}

// NewHUB75Controller creates a new HUB75 controller with the specified configuration
func NewHUB75Controller(config HUB75Config) (*HUB75Controller, error) {
	ctrl := &HUB75Controller{
		config: config,
		lines:  make(map[int]*gpiocdev.Line),
	}
	
	// Request all the GPIO lines
	pins := []int{
		config.R1Pin, config.G1Pin, config.B1Pin,
		config.R2Pin, config.G2Pin, config.B2Pin,
		config.CLKPin, config.OEPin, config.LAPin,
		config.ABPin, config.BCPin, config.CCPin,
		config.DPin, config.EPin,
	}
	
	log.Println("Requesting GPIO lines...")
	for _, pin := range pins {
		line, err := gpiocdev.RequestLine("gpiochip0", pin, gpiocdev.AsOutput(0))
		if err != nil {
			// Clean up any lines we've already requested
			ctrl.Close()
			return nil, err
		}
		ctrl.lines[pin] = line
		log.Printf("Successfully requested GPIO pin %d", pin)
	}
	
	return ctrl, nil
}

// Close releases all GPIO lines
func (c *HUB75Controller) Close() error {
	for pin, line := range c.lines {
		if line != nil {
			if err := line.Close(); err != nil {
				log.Printf("Error closing pin %d: %v", pin, err)
			}
		}
	}
	
	// Clear the map
	c.lines = make(map[int]*gpiocdev.Line)
	return nil
}

// setPin sets the value of a GPIO pin
func (c *HUB75Controller) setPin(pin int, value int) error {
	line, ok := c.lines[pin]
	if !ok {
		return nil // Pin not found, silently ignore
	}
	return line.SetValue(value)
}

// UpdateRow updates a single row of the LED matrix
func (c *HUB75Controller) UpdateRow(rowIdx int, rowData []byte) error {
	// Set address bits based on row index
	addrVal := rowIdx & 0x1F // 5 bits max (A-E)
	
	// Set individual address pins
	if err := c.setPin(c.config.ABPin, (addrVal>>0)&1); err != nil {
		return err
	}
	if err := c.setPin(c.config.BCPin, (addrVal>>1)&1); err != nil {
		return err
	}
	if err := c.setPin(c.config.CCPin, (addrVal>>2)&1); err != nil {
		return err
	}
	if err := c.setPin(c.config.DPin, (addrVal>>3)&1); err != nil {
		return err
	}
	if err := c.setPin(c.config.EPin, (addrVal>>4)&1); err != nil {
		return err
	}
	
	// Disable output during data change
	if err := c.setPin(c.config.OEPin, 1); err != nil {
		return err
	}
	
	// For each pixel in the row
	for col := 0; col < DISPLAY_WIDTH; col++ {
		// Calculate data index for this pixel (6 bytes per pixel)
		idx := col * 6
		
		// Make sure we don't go out of bounds
		if idx+5 >= len(rowData) {
			break
		}
		
		// Set RGB data pins for upper half
		if err := c.setPin(c.config.R1Pin, int(rowData[idx+0])); err != nil {
			return err
		}
		if err := c.setPin(c.config.G1Pin, int(rowData[idx+1])); err != nil {
			return err
		}
		if err := c.setPin(c.config.B1Pin, int(rowData[idx+2])); err != nil {
			return err
		}
		
		// Set RGB data pins for lower half
		if err := c.setPin(c.config.R2Pin, int(rowData[idx+3])); err != nil {
			return err
		}
		if err := c.setPin(c.config.G2Pin, int(rowData[idx+4])); err != nil {
			return err
		}
		if err := c.setPin(c.config.B2Pin, int(rowData[idx+5])); err != nil {
			return err
		}
		
		// Pulse the clock to latch the data
		if err := c.setPin(c.config.CLKPin, 1); err != nil {
			return err
		}
		time.Sleep(time.Microsecond)
		if err := c.setPin(c.config.CLKPin, 0); err != nil {
			return err
		}
	}
	
	// Latch the data
	if err := c.setPin(c.config.LAPin, 1); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if err := c.setPin(c.config.LAPin, 0); err != nil {
		return err
	}
	
	// Enable output
	if err := c.setPin(c.config.OEPin, 0); err != nil {
		return err
	}
	
	return nil
}

// RenderFrame renders a full frame to the LED matrix
func (c *HUB75Controller) RenderFrame(frameData [][]byte) error {
	for rowIdx, rowData := range frameData {
		if err := c.UpdateRow(rowIdx, rowData); err != nil {
			return err
		}
		
		// Small delay between rows to avoid flickering
		time.Sleep(time.Microsecond * 50)
	}
	
	return nil
}

// renderScroll renders text that scrolls across the display
func renderScroll(frameData [][]byte, text string, offset int, color [3]byte) {
	// Create a clean buffer for the new frame
	// This helps prevent flickering and artifacts
	bufferData := make([][]byte, ROWS)
	for i := range bufferData {
		bufferData[i] = make([]byte, DISPLAY_WIDTH*3*2)
	}
	
	// Calculate total text width
	textWidth := 0
	for _, char := range text {
		if _, exists := font5x7[char]; exists {
			textWidth += FONT_WIDTH + CHAR_SPACING
		} else {
			// Space for unknown characters
			textWidth += FONT_WIDTH + CHAR_SPACING
		}
	}
	
	// Starting position (accounting for scroll)
	// This creates a smooth scrolling effect from right to left
	startX := DISPLAY_WIDTH - (offset % (textWidth + DISPLAY_WIDTH))
	
	// Position text vertically centered in the display
	// For a 32-pixel high display with 7-pixel font, center at row 12
	centerY := (DISPLAY_HEIGHT - FONT_HEIGHT) / 2
	
	// Draw the text into the buffer
	cursorX := startX
	for _, char := range text {
		fontData, exists := font5x7[char]
		if !exists {
			fontData = font5x7[' '] // Default to space for unknown characters
		}
		
		// Skip if the character is completely off-screen to the left
		if cursorX + FONT_WIDTH < 0 {
			cursorX += FONT_WIDTH + CHAR_SPACING
			continue
		}
		
		// Skip if the character is completely off-screen to the right
		if cursorX >= DISPLAY_WIDTH {
			break
		}
		
		// Draw the character
		for col := 0; col < FONT_WIDTH; col++ {
			x := cursorX + col
			
			// Skip if this column is off-screen
			if x < 0 || x >= DISPLAY_WIDTH {
				continue
			}
			
			// Get the column data from the font
			colData := byte(0)
			if col < len(fontData) {
				colData = fontData[col]
			}
			
			// Draw each pixel in the column
			for row := 0; row < FONT_HEIGHT; row++ {
				// Check if this pixel is on in the font
				isOn := (colData & (1 << row)) != 0
				
				if isOn {
					// Calculate the actual Y position on the display
					y := centerY + row
					
					// Skip if off-screen vertically
					if y < 0 || y >= DISPLAY_HEIGHT {
						continue
					}
					
					// Map the display coordinates to the appropriate framebuffer location
					// For 32-high displays: Rows 0-15 map to upper half, 16-31 to lower half
					if y < 16 {
						// Upper half - use the R1G1B1 pins (first 3 bytes)
						idx := x * 6
						bufferData[y][idx+0] = color[0] // R1
						bufferData[y][idx+1] = color[1] // G1
						bufferData[y][idx+2] = color[2] // B1
					} else {
						// Lower half - use the R2G2B2 pins (last 3 bytes)
						// Adjust the row to be 0-15 range for the buffer
						idx := x * 6
						bufferData[y-16][idx+3] = color[0] // R2
						bufferData[y-16][idx+4] = color[1] // G2
						bufferData[y-16][idx+5] = color[2] // B2
					}
				}
			}
		}
		
		// Move cursor to the next character position
		cursorX += FONT_WIDTH + CHAR_SPACING
	}
	
	// Copy the buffer to the frame data
	// Only copy once all drawing is complete to prevent flickering
	for row := range bufferData {
		copy(frameData[row], bufferData[row])
	}
}

func main() {
	// Parse command line flags
	textToScroll := flag.String("text", "HELLO WORLD", "Text to scroll across the display")
	showText := flag.Bool("scroll", false, "Show scrolling text instead of test patterns")
	flag.Parse()

	log.Printf("Starting HUB75 display test with scrolling text: %s", *textToScroll)

	// Create HUB75 configuration for Adafruit RGB Matrix Bonnet pinout
	cfg := HUB75Config{
		R1Pin: 5,  // Red data for upper half
		G1Pin: 13, // Green data for upper half
		B1Pin: 6,  // Blue data for upper half
		R2Pin: 12, // Red data for lower half
		G2Pin: 16, // Green data for lower half
		B2Pin: 23, // Blue data for lower half
		CLKPin: 17, // Clock signal
		OEPin: 4,   // Output enable
		LAPin: 21,  // Latch signal
		ABPin: 22,  // Address bit A
		BCPin: 26,  // Address bit B
		CCPin: 27,  // Address bit C
		DPin: 20,   // Address bit D
		EPin: 24,   // Address bit E (for 64-pixel high displays)
	}

	// Initialize HUB75 controller
	hub75, err := NewHUB75Controller(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize HUB75 controller: %v", err)
	}
	defer hub75.Close()

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Prepare frame data
	frameData := make([][]byte, ROWS)
	for i := range frameData {
		// Each row needs RGB data for each pixel
		// For a 64-pixel wide display with two RGB values per pixel (upper/lower):
		// 64 pixels * 3 colors (RGB) * 2 (upper/lower) = 384 bytes per row
		frameData[i] = make([]byte, DISPLAY_WIDTH*3*2)
	}

	// Main display loop
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 50) // 20 FPS for smoother animation
		patternCounter := 0
		scrollOffset := 0
		
		for {
			select {
			case <-sigChan:
				log.Println("Received shutdown signal")
				stop <- struct{}{}
				return
			case <-ticker.C:
				if *showText {
					// Show scrolling text - use red instead of yellow for less brightness
					color := [3]byte{1, 0, 0} // Red text (R only)
					renderScroll(frameData, *textToScroll, scrollOffset, color)
					// Scroll slower for better readability (increment by 1 every frame)
					scrollOffset++
				} else {
					// Show test patterns
					updateFrameData(frameData, patternCounter)
				}
				
				patternCounter++
				
				// Render the frame to the display
				if err := hub75.RenderFrame(frameData); err != nil {
					log.Printf("Error rendering frame: %v", err)
				}
			}
		}
	}()

	<-stop
	log.Println("HUB75 program stopped")
}

// updateFrameData updates the frame data with a test pattern
// patternCounter is used to create animated patterns
func updateFrameData(frameData [][]byte, patternCounter int) {
	switch patternCounter % 4 {
	case 0:
		// All red
		fillColor(frameData, 1, 0, 0)
	case 1:
		// All green
		fillColor(frameData, 0, 1, 0)
	case 2:
		// All blue
		fillColor(frameData, 0, 0, 1)
	case 3:
		// Checkerboard pattern
		fillCheckerboard(frameData, patternCounter)
	}
}

// fillColor fills the entire frame with a solid color
func fillColor(frameData [][]byte, r, g, b byte) {
	// First clear the entire frame
	for row := range frameData {
		for col := 0; col < DISPLAY_WIDTH; col++ {
			idx := col * 6
			for i := 0; i < 6; i++ {
				frameData[row][idx+i] = 0
			}
		}
	}

	// Then set all pixels to the desired color
	for row := range frameData {
		for col := 0; col < DISPLAY_WIDTH; col++ {
			// Set upper half pixel
			upperIdx := col * 6
			frameData[row][upperIdx+0] = r // R1
			frameData[row][upperIdx+1] = g // G1
			frameData[row][upperIdx+2] = b // B1
			
			// Set lower half pixel
			frameData[row][upperIdx+3] = r // R2
			frameData[row][upperIdx+4] = g // G2
			frameData[row][upperIdx+5] = b // B2
		}
	}
}

// fillCheckerboard creates a checkerboard pattern
func fillCheckerboard(frameData [][]byte, offset int) {
	// First clear the entire frame
	for row := range frameData {
		for col := 0; col < DISPLAY_WIDTH; col++ {
			idx := col * 6
			for i := 0; i < 6; i++ {
				frameData[row][idx+i] = 0
			}
		}
	}

	// Create animated checkerboard pattern using the offset for animation
	// Larger cells make the pattern more visible (4x4 pixels per cell)
	cellSize := 4
	for row := range frameData {
		for col := 0; col < DISPLAY_WIDTH; col++ {
			// Determine if this should be an "on" cell in the checkerboard
			// Using integer division to group pixels into cells
			cellRow := row / cellSize
			cellCol := col / cellSize
			isOn := ((cellRow + cellCol + (offset / 8)) % 2) == 0
			
			// Calculate the index in the frame data
			upperIdx := col * 6
			
			// Set both upper and lower half pixels to the same color for consistency
			if isOn {
				// Yellow for "on" cells
				frameData[row][upperIdx+0] = 1 // R1
				frameData[row][upperIdx+1] = 1 // G1
				frameData[row][upperIdx+2] = 0 // B1
				frameData[row][upperIdx+3] = 1 // R2
				frameData[row][upperIdx+4] = 1 // G2
				frameData[row][upperIdx+5] = 0 // B2
			} else {
				// Black for "off" cells
				frameData[row][upperIdx+0] = 0 // R1
				frameData[row][upperIdx+1] = 0 // G1
				frameData[row][upperIdx+2] = 0 // B1
				frameData[row][upperIdx+3] = 0 // R2
				frameData[row][upperIdx+4] = 0 // G2
				frameData[row][upperIdx+5] = 0 // B2
			}
		}
	}
} 