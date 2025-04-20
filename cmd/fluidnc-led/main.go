package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fcurrie/fluidnc-led-golang/internal/config"
	"github.com/fcurrie/fluidnc-led-golang/internal/display"
	"github.com/fcurrie/fluidnc-led-golang/internal/grbl"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create display
	disp, err := display.NewDisplay(cfg.Display)
	if err != nil {
		log.Fatalf("Failed to create display: %v", err)
	}
	defer disp.Close()

	// Create GRBL client
	client, err := grbl.NewClient(cfg.GRBL)
	if err != nil {
		log.Fatalf("Failed to create GRBL client: %v", err)
	}
	defer client.Close()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start GRBL client
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start GRBL client: %v", err)
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")
} 