package gpio

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Pin represents a GPIO pin using the sysfs interface
type Pin struct {
	number int
	mu     sync.Mutex
}

// NewPin creates a new GPIO pin using sysfs
func NewPin(number int) (*Pin, error) {
	log.Printf("Creating GPIO pin %d using sysfs", number)

	// Export the pin
	if err := exportPin(number); err != nil {
		// Ignore "device or resource busy" error, as it might already be exported
		if !os.IsExist(err) && err.Error() != "write /sys/class/gpio/export: device or resource busy" {
			return nil, fmt.Errorf("failed to export pin %d: %v", number, err)
		}
		log.Printf("Pin %d may already be exported, continuing...", number)
	}

	// Short delay to allow sysfs to create the pin directory
	time.Sleep(100 * time.Millisecond)

	// Set the pin as output
	if err := setPinDirection(number, "out"); err != nil {
		return nil, fmt.Errorf("failed to set pin %d direction: %v", number, err)
	}

	return &Pin{
		number: number,
	}, nil
}

// Close closes the GPIO pin
func (p *Pin) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Closing GPIO pin %d", p.number)
	// Unexport the pin
	if err := unexportPin(p.number); err != nil {
		// Ignore errors during unexport, as the pin might not be exported or already cleaned up
		log.Printf("Warning: failed to unexport pin %d: %v", p.number, err)
	}
	return nil
}

// SetValue sets the value of the GPIO pin (0 or 1)
func (p *Pin) SetValue(value int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Setting GPIO pin %d to %d", p.number, value)
	return writePinValue(p.number, value)
}

// GetValue gets the value of the GPIO pin (0 or 1)
func (p *Pin) GetValue() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	value, err := readPinValue(p.number)
	if err != nil {
		return 0, err
	}
	log.Printf("GPIO pin %d value: %d", p.number, value)
	return value, nil
}

// Pulse sends a pulse of the specified duration
func (p *Pin) Pulse(duration time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Sending pulse on GPIO pin %d for %v", p.number, duration)
	if err := writePinValue(p.number, 1); err != nil {
		return err
	}

	time.Sleep(duration)

	return writePinValue(p.number, 0)
}

// Helper functions for sysfs GPIO control

func exportPin(number int) error {
	log.Printf("Exporting GPIO pin %d", number)
	f, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d", number))
	return err
}

func unexportPin(number int) error {
	log.Printf("Unexporting GPIO pin %d", number)
	f, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d", number))
	return err
}

func setPinDirection(number int, direction string) error {
	log.Printf("Setting GPIO pin %d direction to %s", number, direction)
	filePath := fmt.Sprintf("/sys/class/gpio/gpio%d/direction", number)
	f, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", filePath, err)
	}
	defer f.Close()

	_, err = f.WriteString(direction)
	if err != nil {
		return fmt.Errorf("failed to write direction to %s: %v", filePath, err)
	}
	return nil
}

func writePinValue(number int, value int) error {
	log.Printf("Writing value %d to GPIO pin %d", value, number)
	filePath := fmt.Sprintf("/sys/class/gpio/gpio%d/value", number)
	f, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", filePath, err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d", value))
	if err != nil {
		return fmt.Errorf("failed to write value to %s: %v", filePath, err)
	}
	return nil
}

func readPinValue(number int) (int, error) {
	filePath := fmt.Sprintf("/sys/class/gpio/gpio%d/value", number)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open %s: %v", filePath, err)
	}
	defer f.Close()

	var value int
	_, err = fmt.Fscanf(f, "%d", &value)
	if err != nil {
		return 0, fmt.Errorf("failed to read value from %s: %v", filePath, err)
	}
	return value, nil
} 