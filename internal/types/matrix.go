package types

import "image/color"

// Matrix represents a display matrix
type Matrix interface {
	// Clear clears the matrix
	Clear() error
	// SetPixel sets a pixel at the given coordinates to the given color
	SetPixel(x, y int, c color.Color) error
	// Show updates the display with the current buffer
	Show() error
	// Close closes the matrix
	Close() error
} 