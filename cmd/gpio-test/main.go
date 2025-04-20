package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

func main() {
	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting GPIO test...")

	// Try to request GPIO line 5 (BCM GPIO 5) as output
	line, err := gpiocdev.RequestLine("gpiochip0", 5, gpiocdev.AsOutput(0))
	if err != nil {
		log.Printf("Failed to request line: %v", err)
		
		// Try to request a line from gpiochip11 instead (where GPIO base is 512)
		log.Println("Trying with gpiochip11...")
		line, err = gpiocdev.RequestLine("gpiochip11", 512+5, gpiocdev.AsOutput(0))
		if err != nil {
			log.Fatalf("Failed to request line from gpiochip11: %v", err)
		}
	}
	defer line.Close()

	log.Println("Successfully requested GPIO line")

	// Toggle the line every second until terminated
	go func() {
		value := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-sigChan:
				log.Println("Received shutdown signal")
				return
			case <-ticker.C:
				value ^= 1
				if err := line.SetValue(value); err != nil {
					log.Printf("Failed to set value: %v", err)
					continue
				}
				log.Printf("Set GPIO value to %d", value)
			}
		}
	}()

	// Wait for signal to terminate
	<-sigChan
	log.Println("Shutting down...")
} 