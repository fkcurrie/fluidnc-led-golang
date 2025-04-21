package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// Constants for display size
const (
	DISPLAY_WIDTH  = 64  // Width in pixels
	DISPLAY_HEIGHT = 32  // Height in pixels
	FONT_HEIGHT   = 12   // Height of our font in pixels (increased from 7)
	FONT_WIDTH    = 8    // Width of each character in our font (increased from 5)
	CHAR_SPACING  = 2    // Space between characters (increased for readability)
	SCAN_RATE     = 80   // Microseconds per row scan (reduced for faster refresh)
	REFRESH_RATE  = 75   // Frames per second (increased for smoother scrolling)
	SCROLL_SPEED  = 1    // Pixels to move per frame update (reduced for smoother motion)
	FIXED_TIME_PER_FRAME = true // Use fixed timing to prevent flicker
	MIN_BRIGHTNESS = 0.2        // Minimum brightness level to maintain even at low intensity
)

// ComicFont defines a larger 8x12 font with Comic Sans-like rounded styling
var comicFont = map[rune][]byte{
	'A': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000011,
		0b11111111,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b00000000,
		0b00000000,
	},
	'B': {
		0b11111100,
		0b01100110,
		0b01100110,
		0b01100110,
		0b01111100,
		0b01100110,
		0b01100110,
		0b01100110,
		0b01100110,
		0b11111100,
		0b00000000,
		0b00000000,
	},
	'C': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000000,
		0b11000000,
		0b11000000,
		0b11000000,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'D': {
		0b11111000,
		0b01101100,
		0b01100110,
		0b01100011,
		0b01100011,
		0b01100011,
		0b01100011,
		0b01100110,
		0b01101100,
		0b11111000,
		0b00000000,
		0b00000000,
	},
	'E': {
		0b11111110,
		0b01100010,
		0b01100000,
		0b01100000,
		0b01111100,
		0b01100000,
		0b01100000,
		0b01100000,
		0b01100010,
		0b11111110,
		0b00000000,
		0b00000000,
	},
	'F': {
		0b11111110,
		0b01100010,
		0b01100000,
		0b01100000,
		0b01111100,
		0b01100000,
		0b01100000,
		0b01100000,
		0b01100000,
		0b11110000,
		0b00000000,
		0b00000000,
	},
	'G': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000000,
		0b11000000,
		0b11001111,
		0b11000011,
		0b11000011,
		0b01100111,
		0b00111011,
		0b00000000,
		0b00000000,
	},
	'H': {
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11111111,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b00000000,
		0b00000000,
	},
	'I': {
		0b01111100,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b01111100,
		0b00000000,
		0b00000000,
	},
	'J': {
		0b00011110,
		0b00001100,
		0b00001100,
		0b00001100,
		0b00001100,
		0b00001100,
		0b11001100,
		0b11001100,
		0b01101100,
		0b00111000,
		0b00000000,
		0b00000000,
	},
	'K': {
		0b11100111,
		0b01100110,
		0b01100100,
		0b01101000,
		0b01110000,
		0b01111000,
		0b01101100,
		0b01100110,
		0b01100011,
		0b11100001,
		0b00000000,
		0b00000000,
	},
	'L': {
		0b11110000,
		0b01100000,
		0b01100000,
		0b01100000,
		0b01100000,
		0b01100000,
		0b01100000,
		0b01100001,
		0b01100011,
		0b11111111,
		0b00000000,
		0b00000000,
	},
	'M': {
		0b11000011,
		0b11100111,
		0b11111111,
		0b11011011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b00000000,
		0b00000000,
	},
	'N': {
		0b11000011,
		0b11100011,
		0b11110011,
		0b11011011,
		0b11001111,
		0b11000111,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b00000000,
		0b00000000,
	},
	'O': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'P': {
		0b11111100,
		0b01100110,
		0b01100110,
		0b01100110,
		0b01100110,
		0b01111100,
		0b01100000,
		0b01100000,
		0b01100000,
		0b11110000,
		0b00000000,
		0b00000000,
	},
	'Q': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11001011,
		0b11000111,
		0b01100110,
		0b00111101,
		0b00000000,
		0b00000000,
	},
	'R': {
		0b11111100,
		0b01100110,
		0b01100110,
		0b01100110,
		0b01111100,
		0b01101100,
		0b01100110,
		0b01100110,
		0b01100110,
		0b11100110,
		0b00000000,
		0b00000000,
	},
	'S': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b01100000,
		0b00111000,
		0b00001100,
		0b00000110,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'T': {
		0b11111111,
		0b10110110,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b01111000,
		0b00000000,
		0b00000000,
	},
	'U': {
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'V': {
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00011000,
		0b00000000,
		0b00000000,
	},
	'W': {
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11000011,
		0b11011011,
		0b11111111,
		0b01100110,
		0b01100110,
		0b00000000,
		0b00000000,
	},
	'X': {
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00011000,
		0b00011000,
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000011,
		0b00000000,
		0b00000000,
	},
	'Y': {
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00011000,
		0b00011000,
		0b00011000,
		0b00011000,
		0b00011000,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'Z': {
		0b11111111,
		0b11000111,
		0b10001100,
		0b00011000,
		0b00110000,
		0b01100000,
		0b11000000,
		0b11000011,
		0b11100111,
		0b11111111,
		0b00000000,
		0b00000000,
	},
	'0': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000111,
		0b11001111,
		0b11011011,
		0b11110011,
		0b11100011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'1': {
		0b00110000,
		0b01110000,
		0b11110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b11111100,
		0b00000000,
		0b00000000,
	},
	'2': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b00000011,
		0b00000110,
		0b00001100,
		0b00011000,
		0b00110000,
		0b01100000,
		0b11111111,
		0b00000000,
		0b00000000,
	},
	'3': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b00000011,
		0b00011110,
		0b00011110,
		0b00000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'4': {
		0b00001100,
		0b00011100,
		0b00111100,
		0b01101100,
		0b11001100,
		0b11111111,
		0b00001100,
		0b00001100,
		0b00001100,
		0b00011110,
		0b00000000,
		0b00000000,
	},
	'5': {
		0b11111111,
		0b11000000,
		0b11000000,
		0b11000000,
		0b11111100,
		0b00000110,
		0b00000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'6': {
		0b00111100,
		0b01100110,
		0b11000000,
		0b11000000,
		0b11111100,
		0b11000110,
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'7': {
		0b11111111,
		0b11000011,
		0b10000110,
		0b00001100,
		0b00011000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00110000,
		0b00000000,
		0b00000000,
	},
	'8': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000011,
		0b01111110,
		0b01111110,
		0b11000011,
		0b11000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	'9': {
		0b00111100,
		0b01100110,
		0b11000011,
		0b11000011,
		0b01100111,
		0b00111111,
		0b00000011,
		0b00000011,
		0b01100110,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	' ': {
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
	},
	'!': {
		0b00011000,
		0b00111100,
		0b00111100,
		0b00111100,
		0b00111100,
		0b00011000,
		0b00011000,
		0b00000000,
		0b00011000,
		0b00011000,
		0b00000000,
		0b00000000,
	},
	'.': {
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00111100,
		0b00111100,
		0b00000000,
		0b00000000,
	},
	',': {
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00111000,
		0b00111000,
		0b00011000,
		0b00110000,
		0b00000000,
	},
	':': {
		0b00000000,
		0b00000000,
		0b00111100,
		0b00111100,
		0b00000000,
		0b00000000,
		0b00111100,
		0b00111100,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
	},
	'-': {
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b11111111,
		0b11111111,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
		0b00000000,
	},
	'+': {
		0b00000000,
		0b00000000,
		0b00011000,
		0b00011000,
		0b00011000,
		0b11111111,
		0b11111111,
		0b00011000,
		0b00011000,
		0b00011000,
		0b00000000,
		0b00000000,
	},
};

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

// Package-level variables
var (
	isFirstRender = true
	renderLock    sync.Mutex
)

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
	// For Raspberry Pi 5 with pins > 512, we need to use gpiochip0
	chipName := "gpiochip0"
	
	for _, pin := range pins {
		// Adjust GPIO numbers for Pi 5
		adjustedPin := pin - 512
		line, err := gpiocdev.RequestLine(chipName, adjustedPin, gpiocdev.AsOutput(0))
		if err != nil {
			// Clean up any lines we've already requested
			ctrl.Close()
			return nil, err
		}
		ctrl.lines[pin] = line
		log.Printf("Successfully requested GPIO pin %d (adjusted to %d)", pin, adjustedPin)
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

// FrameBuffer represents a full 32-pixel high display buffer
type FrameBuffer struct {
	Pixels [DISPLAY_HEIGHT][DISPLAY_WIDTH][3]byte
}

// NewFrameBuffer creates a new zeroed frame buffer
func NewFrameBuffer() *FrameBuffer {
	return &FrameBuffer{}
}

// Clear zeros out the entire frame buffer
func (fb *FrameBuffer) Clear() {
	for y := 0; y < DISPLAY_HEIGHT; y++ {
		for x := 0; x < DISPLAY_WIDTH; x++ {
			fb.Pixels[y][x][0] = 0 // R
			fb.Pixels[y][x][1] = 0 // G
			fb.Pixels[y][x][2] = 0 // B
		}
	}
}

// SetPixel sets a pixel color at the specified coordinates
func (fb *FrameBuffer) SetPixel(x, y int, r, g, b byte) {
	if x >= 0 && x < DISPLAY_WIDTH && y >= 0 && y < DISPLAY_HEIGHT {
		fb.Pixels[y][x][0] = r
		fb.Pixels[y][x][1] = g
		fb.Pixels[y][x][2] = b
	}
}

// RenderText renders text centered on the display
func (fb *FrameBuffer) RenderText(text string, offsetX int, color [3]byte) {
	fb.Clear()
	
	// Calculate total text width
	textWidth := len(text) * (FONT_WIDTH + CHAR_SPACING)
	
	// Calculate the starting X position with wrapping
	startX := DISPLAY_WIDTH - (offsetX % (textWidth + DISPLAY_WIDTH))
	
	// Calculate vertical position - center the text vertically
	startY := (DISPLAY_HEIGHT - FONT_HEIGHT) / 2
	
	// Draw each character
	x := startX
	for _, char := range text {
		// Skip if the entire character would be off-screen
		if x + FONT_WIDTH < 0 {
			x += FONT_WIDTH + CHAR_SPACING
			continue
		}
		if x >= DISPLAY_WIDTH {
			break
		}
		
		// Get the font data for this character
		fontData, exists := comicFont[char]
		if !exists {
			// Use space for unknown characters
			fontData = comicFont[' ']
		}
		
		// Draw each column of the character
		for col := 0; col < FONT_WIDTH; col++ {
			// Skip if this column is off-screen
			if x + col < 0 || x + col >= DISPLAY_WIDTH {
				continue
			}
			
			// Draw each pixel in the column
			for row := 0; row < FONT_HEIGHT; row++ {
				// Check if this pixel should be on
				if row < len(fontData) && (fontData[row] & (0x80 >> col)) != 0 {
					// Calculate the final position on the display
					displayY := startY + row
					displayX := x + col
					
					// Set pixel
					if displayY >= 0 && displayY < DISPLAY_HEIGHT {
						fb.SetPixel(displayX, displayY, color[0], color[1], color[2])
					}
				}
			}
		}
		
		// Move to the next character position
		x += FONT_WIDTH + CHAR_SPACING
	}
}

// RenderFrame renders a full frame to the LED matrix
func (c *HUB75Controller) RenderFrame(frameBuffer *FrameBuffer) error {
	// On first call, log that we're starting to render
	renderLock.Lock()
	if isFirstRender {
		log.Println("Starting to render frames to the matrix...")
		isFirstRender = false
	}
	renderLock.Unlock()
	
	// Calculate the start time of this frame for consistent timing
	frameStartTime := time.Now()
	targetFrameTime := time.Second / time.Duration(REFRESH_RATE)
	
	// For each row in the 32-pixel high display
	for y := 0; y < DISPLAY_HEIGHT; y++ {
		// Calculate the row address (0-15) and whether this is a top/bottom row
		rowAddress := y % 16
		isBottomHalf := y >= 16
		
		// CRITICAL: Disable output while we set up this row - prevents flickering
		if err := c.setPin(c.config.OEPin, 1); err != nil {
			return err
		}
		
		// Set row address pins (A-E) - fully complete this before moving on
		if err := c.setPin(c.config.ABPin, (rowAddress >> 0) & 1); err != nil { return err }
		if err := c.setPin(c.config.BCPin, (rowAddress >> 1) & 1); err != nil { return err }
		if err := c.setPin(c.config.CCPin, (rowAddress >> 2) & 1); err != nil { return err }
		if err := c.setPin(c.config.DPin, (rowAddress >> 3) & 1); err != nil { return err }
		if err := c.setPin(c.config.EPin, 0); err != nil { return err }
		
		// Pre-clear all RGB pins before setting new values (helps reduce ghosting)
		// Top half clear
		if err := c.setPin(c.config.R1Pin, 0); err != nil { return err }
		if err := c.setPin(c.config.G1Pin, 0); err != nil { return err }
		if err := c.setPin(c.config.B1Pin, 0); err != nil { return err }
		// Bottom half clear
		if err := c.setPin(c.config.R2Pin, 0); err != nil { return err }
		if err := c.setPin(c.config.G2Pin, 0); err != nil { return err }
		if err := c.setPin(c.config.B2Pin, 0); err != nil { return err }
		
		// For each column
		for x := 0; x < DISPLAY_WIDTH; x++ {
			// Get pixel color with intensity correction to avoid flicker at low brightness
			r1, g1, b1 := getAdjustedPixelColor(frameBuffer.Pixels[y][x])
			
			// Set RGB data pins for this pixel
			if isBottomHalf {
				// Bottom half pixels use R2/G2/B2 pins
				if err := c.setPin(c.config.R2Pin, int(r1)); err != nil { return err }
				if err := c.setPin(c.config.G2Pin, int(g1)); err != nil { return err }
				if err := c.setPin(c.config.B2Pin, int(b1)); err != nil { return err }
			} else {
				// Top half pixels use R1/G1/B1 pins
				if err := c.setPin(c.config.R1Pin, int(r1)); err != nil { return err }
				if err := c.setPin(c.config.G1Pin, int(g1)); err != nil { return err }
				if err := c.setPin(c.config.B1Pin, int(b1)); err != nil { return err }
			}
			
			// Clock in this pixel's data - very fast clock for consistent timing
			if err := c.setPin(c.config.CLKPin, 1); err != nil { return err }
			if err := c.setPin(c.config.CLKPin, 0); err != nil { return err }
		}
		
		// CRITICAL: Latch the data to the display drivers
		if err := c.setPin(c.config.LAPin, 1); err != nil { return err }
		if err := c.setPin(c.config.LAPin, 0); err != nil { return err }
		
		// CRITICAL: Enable output only after data is fully latched
		if err := c.setPin(c.config.OEPin, 0); err != nil { return err }
		
		// Wait for scan rate (allows the row to display for the proper amount of time)
		time.Sleep(time.Microsecond * SCAN_RATE)
	}
	
	// If using fixed timing, ensure each frame takes exactly the same amount of time
	if FIXED_TIME_PER_FRAME {
		elapsed := time.Since(frameStartTime)
		if elapsed < targetFrameTime {
			time.Sleep(targetFrameTime - elapsed)
		}
	}
	
	return nil
}

// getAdjustedPixelColor adjusts color intensities to avoid flicker at low brightness
func getAdjustedPixelColor(color [3]byte) (byte, byte, byte) {
	// For each color component, ensure it has at least MIN_BRIGHTNESS if it's on at all
	r, g, b := color[0], color[1], color[2]
	
	// Apply non-linear brightness correction to avoid flicker at low intensities
	// Only apply to non-zero values to maintain true black
	if r > 0 && r < byte(255*MIN_BRIGHTNESS) {
		r = byte(255 * MIN_BRIGHTNESS)
	}
	if g > 0 && g < byte(255*MIN_BRIGHTNESS) {
		g = byte(255 * MIN_BRIGHTNESS)
	}
	if b > 0 && b < byte(255*MIN_BRIGHTNESS) {
		b = byte(255 * MIN_BRIGHTNESS)
	}
	
	return r, g, b
}

func main() {
	// Parse command line flags
	textToScroll := flag.String("text", "HELLO WORLD", "Text to scroll across the display")
	showText := flag.Bool("scroll", false, "Show scrolling text instead of test patterns")
	slowScroll := flag.Bool("slow", false, "Scroll text at a slower speed")
	testMode := flag.Bool("test", false, "Run a simple test pattern only")
	limitRefresh := flag.Int("limit-refresh", 0, "Limit refresh rate to Hz. 0=no limit")
	flag.Parse()

	log.Printf("Starting HUB75 display test with scrolling text: %s", *textToScroll)
	log.Printf("Display configuration: %dx%d pixels", DISPLAY_WIDTH, DISPLAY_HEIGHT)

	// Create HUB75 configuration with Raspberry Pi 5 pins (GPIO base 0)
	// These are the GPIO pin numbers, not the physical pins
	cfg := HUB75Config{
		R1Pin: 5 + 512,   // Red data for upper half
		G1Pin: 13 + 512,  // Green data for upper half
		B1Pin: 6 + 512,   // Blue data for upper half
		R2Pin: 12 + 512,  // Red data for lower half
		G2Pin: 16 + 512,  // Green data for lower half
		B2Pin: 23 + 512,  // Blue data for lower half
		CLKPin: 17 + 512, // Clock signal
		OEPin: 4 + 512,   // Output enable
		LAPin: 21 + 512,  // Latch signal
		ABPin: 22 + 512,  // Address bit A
		BCPin: 26 + 512,  // Address bit B
		CCPin: 27 + 512,  // Address bit C
		DPin:  20 + 512,  // Address bit D
		EPin:  24 + 512,  // Address bit E
	}
	
	log.Printf("GPIO Pin Configuration:")
	log.Printf("R1: %d, G1: %d, B1: %d", cfg.R1Pin-512, cfg.G1Pin-512, cfg.B1Pin-512)
	log.Printf("R2: %d, G2: %d, B2: %d", cfg.R2Pin-512, cfg.G2Pin-512, cfg.B2Pin-512)
	log.Printf("CLK: %d, OE: %d, LA: %d", cfg.CLKPin-512, cfg.OEPin-512, cfg.LAPin-512)
	log.Printf("ROW A: %d, B: %d, C: %d, D: %d, E: %d", 
		cfg.ABPin-512, cfg.BCPin-512, cfg.CCPin-512, cfg.DPin-512, cfg.EPin-512)

	// Initialize HUB75 controller
	hub75, err := NewHUB75Controller(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize HUB75 controller: %v", err)
	}
	defer hub75.Close()
	
	if *testMode {
		// Simple static test pattern
		frameBuffer := NewFrameBuffer()
		
		// Draw test pattern with gradient bars to better detect flickering
		log.Println("Creating gradient test pattern to check for flickering...")
		
		// 1. Clear to black
		for y := 0; y < DISPLAY_HEIGHT; y++ {
			for x := 0; x < DISPLAY_WIDTH; x++ {
				frameBuffer.SetPixel(x, y, 0, 0, 0)
			}
		}
		
		// 2. Draw horizontal gradient bars - red, green, blue
		barHeight := DISPLAY_HEIGHT / 3
		
		// Red gradient (top)
		for y := 0; y < barHeight; y++ {
			for x := 0; x < DISPLAY_WIDTH; x++ {
				intensity := byte((x * 255) / DISPLAY_WIDTH)
				frameBuffer.SetPixel(x, y, intensity, 0, 0)
			}
		}
		
		// Green gradient (middle)
		for y := barHeight; y < barHeight*2; y++ {
			for x := 0; x < DISPLAY_WIDTH; x++ {
				intensity := byte((x * 255) / DISPLAY_WIDTH)
				frameBuffer.SetPixel(x, y, 0, intensity, 0)
			}
		}
		
		// Blue gradient (bottom)
		for y := barHeight*2; y < DISPLAY_HEIGHT; y++ {
			for x := 0; x < DISPLAY_WIDTH; x++ {
				intensity := byte((x * 255) / DISPLAY_WIDTH)
				frameBuffer.SetPixel(x, y, 0, 0, intensity)
			}
		}
		
		// Render test pattern for 10 seconds
		log.Println("Rendering test pattern for 10 seconds...")
		startTime := time.Now()
		for time.Since(startTime) < 10*time.Second {
			if err := hub75.RenderFrame(frameBuffer); err != nil {
				log.Printf("Error rendering test frame: %v", err)
				break
			}
		}
		
		log.Println("Test pattern complete.")
		return
	}

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	stop := make(chan struct{})

	// Main display loop
	go func() {
		// Double buffering
		frameBuffer1 := NewFrameBuffer()
		frameBuffer2 := NewFrameBuffer()
		
		// Current display buffer and next buffer
		displayBuffer := frameBuffer1
		nextBuffer := frameBuffer2
		
		// For smooth scrolling
		scrollOffset := 0
		frameCounter := 0  // For tracking animation frames
		
		// Fixed frame rate ticker with limiter if specified
		frameRate := REFRESH_RATE
		if *limitRefresh > 0 && *limitRefresh < REFRESH_RATE {
			frameRate = *limitRefresh
		}
		frameTicker := time.NewTicker(time.Second / time.Duration(frameRate))
		
		// Initialize both buffers with the same content
		color := [3]byte{1, 0, 0} // Red text
		displayBuffer.RenderText(*textToScroll, scrollOffset, color)
		nextBuffer.RenderText(*textToScroll, scrollOffset, color)
		
		// Render loop
		for {
			select {
			case <-sigChan:
				log.Println("Received shutdown signal")
				stop <- struct{}{}
				return
			case <-frameTicker.C:
				// Render current frame
				if err := hub75.RenderFrame(displayBuffer); err != nil {
					log.Printf("Error rendering frame: %v", err)
				}
				
				// Update scroll offset for next frame
				if *showText {
					// Update scrolling speed based on slow flag
					speed := SCROLL_SPEED
					if *slowScroll {
						// Still move but at reduced speed
						if frameCounter % 5 == 0 {
							scrollOffset += 1
						}
					} else {
						scrollOffset += speed
					}
					
					// Prepare next buffer
					nextBuffer.RenderText(*textToScroll, scrollOffset, color)
				} else {
					// For non-scrolling modes, update color pattern
					pattern := frameCounter % 3
					var r, g, b byte
					switch pattern {
					case 0:
						r, g, b = 1, 0, 0 // Red
					case 1:
						r, g, b = 0, 1, 0 // Green
					case 2:
						r, g, b = 0, 0, 1 // Blue
					}
					
					// Fill display with solid color
					for y := 0; y < DISPLAY_HEIGHT; y++ {
						for x := 0; x < DISPLAY_WIDTH; x++ {
							nextBuffer.SetPixel(x, y, r, g, b)
						}
					}
				}
				
				// Swap buffers for next frame
				displayBuffer, nextBuffer = nextBuffer, displayBuffer
				
				// Track frames for animation timing
				frameCounter++
			}
		}
	}()

	<-stop
	log.Println("HUB75 program stopped")
} 