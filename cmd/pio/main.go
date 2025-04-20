package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fcurrie/fluidnc-led-golang/pkg/pio"
)

// Constants for display size
const (
	DISPLAY_WIDTH  = 32 // Width in pixels for a standard 32x32 panel
	DISPLAY_HEIGHT = 32 // Height in pixels for a standard 32x32 panel
	ROWS          = 16  // Number of addressable rows (DISPLAY_HEIGHT/2 for panels with upper/lower half data)
)

func main() {
	// Parse command line flags
	pioNum := flag.Int("pio", 0, "PIO number (0-1)")
	smNum := flag.Int("sm", 0, "State machine number (0-3)")
	flag.Parse()

	log.Printf("Starting HUB75 display test with PIO%d SM%d", *pioNum, *smNum)

	// Initialize PIO
	p, err := pio.NewPIO()
	if err != nil {
		log.Fatalf("Failed to initialize PIO: %v", err)
	}
	defer p.Close()

	// Create HUB75 program configuration - using Adafruit RGB Matrix Bonnet pinout
	cfg := pio.HUB75Program{
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

	// Initialize HUB75 program
	hub75, err := pio.NewHUB75Program(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize HUB75 program: %v", err)
	}
	defer hub75.Close()

	// Get program and pins from HUB75 configuration
	program := hub75.GetProgram()
	pins := hub75.GetPins()

	// Initialize state machine with HUB75 program and pins
	sm, err := pio.NewStateMachine(pio.Config{
		ChipNumber: "gpiochip0", // Use gpiochip0 for Raspberry Pi 5
		SMNumber:   *smNum,
		Program:    program,
		Pins:       pins,
	})
	if err != nil {
		log.Fatalf("Failed to initialize state machine: %v", err)
	}
	defer sm.Close()

	// Load the HUB75 program
	if err := hub75.LoadProgram(sm); err != nil {
		log.Fatalf("Failed to load HUB75 program: %v", err)
	}

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the HUB75 program
	if err := hub75.Start(sm); err != nil {
		log.Fatalf("Failed to start HUB75 program: %v", err)
	}
	log.Println("HUB75 program started")

	// Prepare frame data
	frameData := make([][]byte, ROWS)
	for i := range frameData {
		// Each row needs RGB data for each pixel
		// For a 32-pixel wide display with two RGB values per pixel (upper/lower):
		// 32 pixels * 3 colors (RGB) * 2 (upper/lower) = 192 bytes per row
		frameData[i] = make([]byte, DISPLAY_WIDTH*3*2)
	}

	// Main display loop
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100) // 10 FPS
		patternCounter := 0
		
		for {
			select {
			case <-sigChan:
				log.Println("Received shutdown signal")
				stop <- struct{}{}
				return
			case <-ticker.C:
				// Update the frame data with a new pattern
				updateFrameData(frameData, patternCounter)
				patternCounter++
				
				// Render the frame to the display
				if err := hub75.RenderFrame(sm, frameData); err != nil {
					log.Printf("Error rendering frame: %v", err)
				}
			}
		}
	}()

	<-stop

	// Stop the HUB75 program
	if err := hub75.Stop(sm); err != nil {
		log.Printf("Error stopping HUB75 program: %v", err)
	}
	log.Println("HUB75 program stopped")
}

// updateFrameData updates the frame data with a test pattern
// patternCounter is used to create animated patterns
func updateFrameData(frameData [][]byte, patternCounter int) {
	switch patternCounter % 4 {
	case 0:
		// All red
		fillColor(frameData, 255, 0, 0)
	case 1:
		// All green
		fillColor(frameData, 0, 255, 0)
	case 2:
		// All blue
		fillColor(frameData, 0, 0, 255)
	case 3:
		// Checkerboard pattern
		fillCheckerboard(frameData, patternCounter)
	}
}

// fillColor fills the entire frame with a solid color
func fillColor(frameData [][]byte, r, g, b byte) {
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
	for row := range frameData {
		for col := 0; col < DISPLAY_WIDTH; col++ {
			upperIdx := col * 6
			
			// Determine if this should be an "on" cell in the checkerboard
			isOn := ((row + col + offset) % 2) == 0
			
			// Set upper half pixel
			if isOn {
				frameData[row][upperIdx+0] = 255 // R1
				frameData[row][upperIdx+1] = 255 // G1
				frameData[row][upperIdx+2] = 0   // B1
			} else {
				frameData[row][upperIdx+0] = 0   // R1
				frameData[row][upperIdx+1] = 0   // G1
				frameData[row][upperIdx+2] = 0   // B1
			}
			
			// Set lower half pixel
			if isOn {
				frameData[row][upperIdx+3] = 255 // R2
				frameData[row][upperIdx+4] = 0   // G2
				frameData[row][upperIdx+5] = 255 // B2
			} else {
				frameData[row][upperIdx+3] = 0   // R2
				frameData[row][upperIdx+4] = 0   // G2
				frameData[row][upperIdx+5] = 0   // B2
			}
		}
	}
} 