package pio

import (
	"fmt"
	"sync"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// HUB75Program represents a PIO program for HUB75 LED matrices using Adafruit RGB Matrix Bonnet
type HUB75Program struct {
	// Pin definitions for Adafruit RGB Matrix Bonnet
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
	
	// Private fields
	mu      sync.Mutex
	lines   map[int]*gpiocdev.Line // Map of GPIO pin numbers to Line objects
}

// NewHUB75Program creates a new HUB75 program with the Adafruit RGB Matrix Bonnet pin configuration
func NewHUB75Program(cfg HUB75Program) (*HUB75Program, error) {
	// Validate pins
	if cfg.R1Pin < 0 || cfg.G1Pin < 0 || cfg.B1Pin < 0 ||
		cfg.R2Pin < 0 || cfg.G2Pin < 0 || cfg.B2Pin < 0 ||
		cfg.CLKPin < 0 || cfg.OEPin < 0 || cfg.LAPin < 0 ||
		cfg.ABPin < 0 || cfg.BCPin < 0 || cfg.CCPin < 0 ||
		cfg.DPin < 0 || cfg.EPin < 0 {
		return nil, fmt.Errorf("invalid pin configuration: all pins must be non-negative")
	}

	return &HUB75Program{
		R1Pin:  cfg.R1Pin,
		G1Pin:  cfg.G1Pin,
		B1Pin:  cfg.B1Pin,
		R2Pin:  cfg.R2Pin,
		G2Pin:  cfg.G2Pin,
		B2Pin:  cfg.B2Pin,
		CLKPin: cfg.CLKPin,
		OEPin:  cfg.OEPin,
		LAPin:  cfg.LAPin,
		ABPin:  cfg.ABPin,
		BCPin:  cfg.BCPin,
		CCPin:  cfg.CCPin,
		DPin:   cfg.DPin,
		EPin:   cfg.EPin,
		lines:  make(map[int]*gpiocdev.Line),
	}, nil
}

// GetProgram returns the PIO program for HUB75 using Adafruit RGB Matrix Bonnet
// This is based on the Adafruit Blinka Raspberry Pi 5 Piomatter implementation
func (p *HUB75Program) GetProgram() []uint16 {
	/*
	   Implementation based on Adafruit's PIO assembly for HUB75:
	   
	   .program hub75
	   .side_set 1
	   
	   loop:
	       out pins, 6   side 0 ; Output R1,G1,B1,R2,G2,B2 data, clock low
	       nop           side 1 ; Clock high (data latched by panel)
	       jmp loop      side 0 ; Clock low, loop back
	*/
	
	// Direct translation of the assembly above to PIO machine code
	// Format of instructions:
	// - Bits 0-4: Destination (pins)
	// - Bits 5-9: Operation data (shift count = 6)
	// - Bits 10-12: Source (OUT instruction = 011)
	// - Bits 13-14: Delay (0)
	// - Bit 15: Side-set enable
	return []uint16{
		0x6003, // OUT pins, 6      side 0  -- Send 6 bits to pins, clock low
		0xA042, // NOP              side 1  -- Clock high (data latched)
		0x0001, // JMP loop         side 0  -- Clock low, loop back
	}
}

// GetPins returns the pins used by the HUB75 program on Adafruit RGB Matrix Bonnet
func (p *HUB75Program) GetPins() []int {
	return []int{
		p.R1Pin, p.G1Pin, p.B1Pin,
		p.R2Pin, p.G2Pin, p.B2Pin,
		p.CLKPin, p.OEPin, p.LAPin,
		p.ABPin, p.BCPin, p.CCPin,
		p.DPin, p.EPin,
	}
}

// LoadProgram loads the HUB75 program into the PIO state machine
func (p *HUB75Program) LoadProgram(sm *StateMachine) error {
	if sm == nil {
		return fmt.Errorf("state machine is nil")
	}
	
	// Configure all pins for output
	pins := p.GetPins()
	for _, pin := range pins {
		if err := sm.ConfigurePin(pin); err != nil {
			return fmt.Errorf("failed to configure pin %d: %v", pin, err)
		}
	}

	// Load the PIO program
	program := p.GetProgram()
	for i, instr := range program {
		if err := sm.pio.writeReg(PIOBaseAddr+uint32(i*4), uint32(instr)); err != nil {
			return fmt.Errorf("failed to write instruction %d: %v", i, err)
		}
	}

	// Configure the state machine for HUB75
	// Set pin group: R1,G1,B1,R2,G2,B2 as OUT pins
	// Set CLK as side-set pin
	pinCtrl := uint32(0)
	pinCtrl |= uint32(p.R1Pin) // OUT base pin = R1
	pinCtrl |= uint32(5) << 20 // OUT count = 6 pins (n-1)
	pinCtrl |= uint32(p.CLKPin) << 10 // Side-set base = CLK
	pinCtrl |= uint32(0) << 12 // Side-set count = 1 pin (n-1)
	
	smOffset := uint32(sm.sm * SM_OFFSET)
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_PINCTRL, pinCtrl); err != nil {
		return fmt.Errorf("failed to configure pin control: %v", err)
	}

	return nil
}

// Start begins the HUB75 display operation
func (p *HUB75Program) Start(sm *StateMachine) error {
	if sm == nil {
		return fmt.Errorf("state machine is nil")
	}
	return sm.Start()
}

// Stop halts the HUB75 display operation
func (p *HUB75Program) Stop(sm *StateMachine) error {
	if sm == nil {
		return fmt.Errorf("state machine is nil")
	}
	return sm.Stop()
}

// Close releases all resources used by the HUB75 program
func (p *HUB75Program) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Close any GPIO lines that we opened
	for _, line := range p.lines {
		if line != nil {
			line.Close()
		}
	}
	
	// Clear the lines map
	p.lines = make(map[int]*gpiocdev.Line)
	
	return nil
}

// getOrRequestLine gets an existing line or requests a new one
func (p *HUB75Program) getOrRequestLine(sm *StateMachine, pin int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if _, exists := p.lines[pin]; !exists {
		// Request the line from the state machine's chip
		line, err := sm.chip.RequestLine(pin, gpiocdev.AsOutput(0))
		if err != nil {
			return fmt.Errorf("failed to request line for pin %d: %v", pin, err)
		}
		p.lines[pin] = line
	}
	
	return nil
}

// setPin sets the value of a GPIO pin
func (p *HUB75Program) setPin(pin int, value int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	line, exists := p.lines[pin]
	if !exists {
		return fmt.Errorf("pin %d not configured", pin)
	}
	
	return line.SetValue(value)
}

// UpdateRow updates a single row of the LED matrix
// This function handles the address bits and data output
func (p *HUB75Program) UpdateRow(sm *StateMachine, rowIdx int, rowData []byte) error {
	if sm == nil {
		return fmt.Errorf("state machine is nil")
	}
	
	// Ensure we have lines for all the pins we need
	pins := []int{p.ABPin, p.BCPin, p.CCPin, p.DPin, p.EPin, p.OEPin, p.LAPin}
	for _, pin := range pins {
		if err := p.getOrRequestLine(sm, pin); err != nil {
			return err
		}
	}
	
	// Set address bits based on row index
	addrVal := rowIdx & 0x1F // 5 bits max (A-E)
	
	// Set individual address pins
	if err := p.setPin(p.ABPin, (addrVal>>0)&1); err != nil {
		return fmt.Errorf("failed to set address bit A: %v", err)
	}
	if err := p.setPin(p.BCPin, (addrVal>>1)&1); err != nil {
		return fmt.Errorf("failed to set address bit B: %v", err)
	}
	if err := p.setPin(p.CCPin, (addrVal>>2)&1); err != nil {
		return fmt.Errorf("failed to set address bit C: %v", err)
	}
	if err := p.setPin(p.DPin, (addrVal>>3)&1); err != nil {
		return fmt.Errorf("failed to set address bit D: %v", err)
	}
	if err := p.setPin(p.EPin, (addrVal>>4)&1); err != nil {
		return fmt.Errorf("failed to set address bit E: %v", err)
	}
	
	// Disable output during data change
	if err := p.setPin(p.OEPin, 1); err != nil {
		return fmt.Errorf("failed to disable output: %v", err)
	}
	
	// For each pixel in the row, send RGB data
	for i := 0; i < len(rowData); i += 6 {
		// Pack data for upper and lower half of the panel in 6-bit format:
		// R1, G1, B1, R2, G2, B2
		if i+5 < len(rowData) {
			data := uint32(0)
			if rowData[i+0] > 0 {
				data |= 1 << 0 // R1
			}
			if rowData[i+1] > 0 {
				data |= 1 << 1 // G1
			}
			if rowData[i+2] > 0 {
				data |= 1 << 2 // B1
			}
			if rowData[i+3] > 0 {
				data |= 1 << 3 // R2
			}
			if rowData[i+4] > 0 {
				data |= 1 << 4 // G2
			}
			if rowData[i+5] > 0 {
				data |= 1 << 5 // B2
			}
			
			// Send data to the state machine
			if err := sm.Put(data); err != nil {
				return fmt.Errorf("failed to send pixel data: %v", err)
			}
		}
	}
	
	// Latch the data
	if err := p.setPin(p.LAPin, 1); err != nil {
		return fmt.Errorf("failed to set latch high: %v", err)
	}
	
	// Small delay to ensure latch is processed
	time.Sleep(time.Microsecond)
	
	if err := p.setPin(p.LAPin, 0); err != nil {
		return fmt.Errorf("failed to set latch low: %v", err)
	}
	
	// Enable output
	if err := p.setPin(p.OEPin, 0); err != nil {
		return fmt.Errorf("failed to enable output: %v", err)
	}
	
	return nil
}

// RenderFrame renders a full frame to the LED matrix
// The frameData should be a 2D array of RGB values [rows][columns*3]
func (p *HUB75Program) RenderFrame(sm *StateMachine, frameData [][]byte) error {
	if sm == nil {
		return fmt.Errorf("state machine is nil")
	}
	
	for rowIdx, rowData := range frameData {
		if err := p.UpdateRow(sm, rowIdx, rowData); err != nil {
			return fmt.Errorf("failed to update row %d: %v", rowIdx, err)
		}
		
		// Small delay between rows to avoid flickering
		time.Sleep(time.Microsecond * 50)
	}
	
	return nil
} 