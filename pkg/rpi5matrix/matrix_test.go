package rpi5matrix

import (
	"image/color"
	"testing"
)

// TestNewMatrix tests the creation of a new matrix
func TestNewMatrix(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Width:      32,
				Height:     8,
				Brightness: 128,
				GPIOPin:    18,
			},
			wantErr: false,
		},
		{
			name: "invalid width",
			cfg: &Config{
				Width:      0,
				Height:     8,
				Brightness: 128,
				GPIOPin:    18,
			},
			wantErr: true,
		},
		{
			name: "invalid height",
			cfg: &Config{
				Width:      32,
				Height:     0,
				Brightness: 128,
				GPIOPin:    18,
			},
			wantErr: true,
		},
		{
			name: "invalid brightness",
			cfg: &Config{
				Width:      32,
				Height:     8,
				Brightness: 256,
				GPIOPin:    18,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix, err := NewMatrix(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMatrix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && matrix == nil {
				t.Error("NewMatrix() returned nil matrix when no error expected")
			}
		})
	}
}

// TestMatrixOperations tests basic matrix operations
func TestMatrixOperations(t *testing.T) {
	cfg := &Config{
		Width:      32,
		Height:     8,
		Brightness: 128,
		GPIOPin:    18,
	}

	matrix, err := NewMatrix(cfg)
	if err != nil {
		t.Fatalf("Failed to create matrix: %v", err)
	}
	defer matrix.Close()

	// Test dimensions
	width, height := matrix.GetDimensions()
	if width != cfg.Width || height != cfg.Height {
		t.Errorf("GetDimensions() = %dx%d, want %dx%d", width, height, cfg.Width, cfg.Height)
	}

	// Test brightness
	brightness := matrix.GetBrightness()
	if brightness != cfg.Brightness {
		t.Errorf("GetBrightness() = %d, want %d", brightness, cfg.Brightness)
	}

	// Test setting brightness
	newBrightness := 64
	if err := matrix.SetBrightness(newBrightness); err != nil {
		t.Errorf("SetBrightness() error = %v", err)
	}
	brightness = matrix.GetBrightness()
	if brightness != newBrightness {
		t.Errorf("GetBrightness() after SetBrightness() = %d, want %d", brightness, newBrightness)
	}

	// Test pixel operations
	red := color.RGBA{255, 0, 0, 255}
	if err := matrix.SetPixel(0, 0, red); err != nil {
		t.Errorf("SetPixel() error = %v", err)
	}

	r, g, b, err := matrix.GetPixelColor(0, 0)
	if err != nil {
		t.Errorf("GetPixelColor() error = %v", err)
	}
	if r != 255 || g != 0 || b != 0 {
		t.Errorf("GetPixelColor() = (%d, %d, %d), want (255, 0, 0)", r, g, b)
	}

	// Test HSV color
	if err := matrix.SetPixelHSV(1, 1, 0, 1, 1); err != nil {
		t.Errorf("SetPixelHSV() error = %v", err)
	}

	r, g, b, err = matrix.GetPixelColor(1, 1)
	if err != nil {
		t.Errorf("GetPixelColor() error = %v", err)
	}
	if r != 255 || g != 0 || b != 0 {
		t.Errorf("GetPixelColor() after SetPixelHSV() = (%d, %d, %d), want (255, 0, 0)", r, g, b)
	}

	// Test clear
	if err := matrix.Clear(); err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// Test out of bounds
	if err := matrix.SetPixel(-1, 0, red); err == nil {
		t.Error("SetPixel() with negative x did not return error")
	}
	if err := matrix.SetPixel(0, -1, red); err == nil {
		t.Error("SetPixel() with negative y did not return error")
	}
	if err := matrix.SetPixel(cfg.Width, 0, red); err == nil {
		t.Error("SetPixel() with x >= width did not return error")
	}
	if err := matrix.SetPixel(0, cfg.Height, red); err == nil {
		t.Error("SetPixel() with y >= height did not return error")
	}
}

// TestMatrixConcurrency tests concurrent access to the matrix
func TestMatrixConcurrency(t *testing.T) {
	cfg := &Config{
		Width:      32,
		Height:     8,
		Brightness: 128,
		GPIOPin:    18,
	}

	matrix, err := NewMatrix(cfg)
	if err != nil {
		t.Fatalf("Failed to create matrix: %v", err)
	}
	defer matrix.Close()

	// Test concurrent pixel setting
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				x := (i + j) % cfg.Width
				y := (i + j) % cfg.Height
				err := matrix.SetPixel(x, y, color.RGBA{uint8(i), uint8(j), 0, 255})
				if err != nil {
					t.Errorf("SetPixel() error = %v", err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
} 