package main

import (
	"log"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

func main() {
	// Try to access a single GPIO pin to verify functionality
	chipName := "gpiochip0"
	
	// Try a simple pin to test (BCM GPIO 2)
	pin := 2
	
	log.Printf("Testing GPIO pin %d on chip %s", pin, chipName)
	
	// Request the GPIO line
	line, err := gpiocdev.RequestLine(chipName, pin, gpiocdev.AsOutput(0))
	if err != nil {
		log.Fatalf("Failed to request GPIO line: %v", err)
	}
	defer line.Close()
	
	log.Println("Successfully requested GPIO line, blinking 10 times...")
	
	// Blink the pin 10 times
	for i := 0; i < 10; i++ {
		if err := line.SetValue(1); err != nil {
			log.Fatalf("Failed to set GPIO high: %v", err)
		}
		log.Printf("Set GPIO pin %d HIGH", pin)
		time.Sleep(time.Second)
		
		if err := line.SetValue(0); err != nil {
			log.Fatalf("Failed to set GPIO low: %v", err)
		}
		log.Printf("Set GPIO pin %d LOW", pin)
		time.Sleep(time.Second)
	}
	
	log.Println("GPIO test completed successfully")
} 