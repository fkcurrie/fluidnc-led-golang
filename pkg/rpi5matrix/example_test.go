package rpi5matrix_test

import (
	"fmt"
	"image/color"
	"time"

	"github.com/fkcurrie/fluidnc-led-golang/pkg/rpi5matrix"
)

func Example() {
	// Create a new matrix with default configuration
	cfg := &rpi5matrix.Config{
		Width:      32,
		Height:     8,
		Brightness: 128,
		GPIOPin:    18,
	}

	matrix, err := rpi5matrix.NewMatrix(cfg)
	if err != nil {
		fmt.Printf("Failed to create matrix: %v\n", err)
		return
	}
	defer matrix.Close()

	// Clear the matrix
	if err := matrix.Clear(); err != nil {
		fmt.Printf("Failed to clear matrix: %v\n", err)
		return
	}

	// Set some pixels
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Red
		color.RGBA{0, 255, 0, 255},   // Green
		color.RGBA{0, 0, 255, 255},   // Blue
		color.RGBA{255, 255, 0, 255}, // Yellow
	}

	for i, c := range colors {
		if err := matrix.SetPixel(i, 0, c); err != nil {
			fmt.Printf("Failed to set pixel: %v\n", err)
			return
		}
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}

	// Wait a bit
	time.Sleep(time.Second)

	// Set brightness
	if err := matrix.SetBrightness(64); err != nil {
		fmt.Printf("Failed to set brightness: %v\n", err)
		return
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}

	// Wait a bit
	time.Sleep(time.Second)

	// Clear the matrix
	if err := matrix.Clear(); err != nil {
		fmt.Printf("Failed to clear matrix: %v\n", err)
		return
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}
}

func ExampleMatrix_SetPixelHSV() {
	cfg := &rpi5matrix.Config{
		Width:      32,
		Height:     8,
		Brightness: 128,
		GPIOPin:    18,
	}

	matrix, err := rpi5matrix.NewMatrix(cfg)
	if err != nil {
		fmt.Printf("Failed to create matrix: %v\n", err)
		return
	}
	defer matrix.Close()

	// Set a pixel using HSV color
	// Hue: 0 (red), Saturation: 1 (full), Value: 1 (full)
	if err := matrix.SetPixelHSV(0, 0, 0, 1, 1); err != nil {
		fmt.Printf("Failed to set pixel: %v\n", err)
		return
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}
}

func ExampleMatrix_SetText() {
	cfg := &rpi5matrix.Config{
		Width:      32,
		Height:     8,
		Brightness: 128,
		GPIOPin:    18,
	}

	matrix, err := rpi5matrix.NewMatrix(cfg)
	if err != nil {
		fmt.Printf("Failed to create matrix: %v\n", err)
		return
	}
	defer matrix.Close()

	// Set text on the matrix
	if err := matrix.SetText("Hello", 0, 0, color.RGBA{255, 255, 255, 255}); err != nil {
		fmt.Printf("Failed to set text: %v\n", err)
		return
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}
}

func ExampleMatrix_Scroll() {
	cfg := &rpi5matrix.Config{
		Width:      32,
		Height:     8,
		Brightness: 128,
		GPIOPin:    18,
	}

	matrix, err := rpi5matrix.NewMatrix(cfg)
	if err != nil {
		fmt.Printf("Failed to create matrix: %v\n", err)
		return
	}
	defer matrix.Close()

	// Set some pixels
	for i := 0; i < cfg.Width; i++ {
		if err := matrix.SetPixel(i, 0, color.RGBA{255, 0, 0, 255}); err != nil {
			fmt.Printf("Failed to set pixel: %v\n", err)
			return
		}
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}

	// Scroll the display
	if err := matrix.Scroll(-1, 0); err != nil {
		fmt.Printf("Failed to scroll matrix: %v\n", err)
		return
	}

	// Show the changes
	if err := matrix.Show(); err != nil {
		fmt.Printf("Failed to show matrix: %v\n", err)
		return
	}
} 