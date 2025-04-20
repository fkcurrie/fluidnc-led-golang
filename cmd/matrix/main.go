package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/fkcurrie/fluidnc-led-golang/internal/config"
	"github.com/fkcurrie/fluidnc-led-golang/pkg/rpi5matrix"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Failed to load config from %s: %v", *configPath, err)
		log.Printf("Using default configuration")
		cfg = config.DefaultConfig()
	}

	// Create matrix configuration
	matrixCfg := &rpi5matrix.Config{
		Width:      cfg.Display.Width,
		Height:     cfg.Display.Height,
		Brightness: cfg.Display.Brightness,
		GPIOPin:    530, // GPIO 18 on Raspberry Pi 5 is actually GPIO 530
	}

	// Create matrix
	matrix, err := rpi5matrix.NewMatrix(matrixCfg)
	if err != nil {
		log.Fatalf("Failed to create matrix: %v", err)
	}
	defer matrix.Close()

	// Test pattern 1: All red
	log.Println("Setting all pixels to red")
	for i := 0; i < matrixCfg.Width*matrixCfg.Height; i++ {
		if err := matrix.SetPixel(i, 0, color.RGBA{255, 0, 0, 255}); err != nil {
			log.Fatalf("Failed to set pixel: %v", err)
		}
	}
	if err := matrix.Show(); err != nil {
		log.Fatalf("Failed to show matrix: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Test pattern 2: All green
	log.Println("Setting all pixels to green")
	for i := 0; i < matrixCfg.Width*matrixCfg.Height; i++ {
		if err := matrix.SetPixel(i, 0, color.RGBA{0, 255, 0, 255}); err != nil {
			log.Fatalf("Failed to set pixel: %v", err)
		}
	}
	if err := matrix.Show(); err != nil {
		log.Fatalf("Failed to show matrix: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Test pattern 3: All blue
	log.Println("Setting all pixels to blue")
	for i := 0; i < matrixCfg.Width*matrixCfg.Height; i++ {
		if err := matrix.SetPixel(i, 0, color.RGBA{0, 0, 255, 255}); err != nil {
			log.Fatalf("Failed to set pixel: %v", err)
		}
	}
	if err := matrix.Show(); err != nil {
		log.Fatalf("Failed to show matrix: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Test pattern 4: Alternating pixels
	log.Println("Setting alternating pixels")
	for i := 0; i < matrixCfg.Width*matrixCfg.Height; i++ {
		if i%2 == 0 {
			if err := matrix.SetPixel(i, 0, color.RGBA{255, 255, 255, 255}); err != nil {
				log.Fatalf("Failed to set pixel: %v", err)
			}
		} else {
			if err := matrix.SetPixel(i, 0, color.RGBA{0, 0, 0, 255}); err != nil {
				log.Fatalf("Failed to set pixel: %v", err)
			}
		}
	}
	if err := matrix.Show(); err != nil {
		log.Fatalf("Failed to show matrix: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Clear the matrix
	log.Println("Clearing matrix")
	if err := matrix.Clear(); err != nil {
		log.Fatalf("Failed to clear matrix: %v", err)
	}
	if err := matrix.Show(); err != nil {
		log.Fatalf("Failed to show matrix: %v", err)
	}

	fmt.Println("Test completed successfully")
} 