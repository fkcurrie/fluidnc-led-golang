package pio

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/warthog618/go-gpiocdev"
	"golang.org/x/sys/unix"
)

const (
	// GPIO base address for Raspberry Pi 5
	GPIOBase = 0xfe200000

	// GPIO pin numbers for HUB75 interface (adjusted for PIO)
	R1_PIN = 0  // Red data for upper half
	G1_PIN = 1  // Green data for upper half
	B1_PIN = 2  // Blue data for upper half
	R2_PIN = 3  // Red data for lower half
	G2_PIN = 4  // Green data for lower half
	B2_PIN = 5  // Blue data for lower half
	A_PIN  = 6  // Row address bit A
	B_PIN  = 7  // Row address bit B
	C_PIN  = 8  // Row address bit C
	CLK_PIN = 9  // Clock
	LAT_PIN = 10 // Latch
	OE_PIN  = 11 // Output enable

	// PIO base address for Raspberry Pi 5 (RP1)
	PIOBaseAddr = 0x50200000

	// PIO memory size (4KB per PIO block)
	PIOMemSize = 0x1000

	// PIO register offsets
	SM0_CLKDIV    = 0x0c8
	SM0_EXECCTRL  = 0x0cc
	SM0_SHIFTCTRL = 0x0d0
	SM0_ADDR      = 0x0d4
	SM0_INSTR     = 0x0d8
	SM0_PINCTRL   = 0x0dc
	SM0_FSTAT     = 0x0e0
	SM0_RXF       = 0x0e4
	SM0_TXF       = 0x0e8

	// State machine offset
	SM_OFFSET = 0x024
)

// PIO represents a PIO controller
type PIO struct {
	mu sync.Mutex
	chip *gpiocdev.Chip
	pio *os.File
	mem []byte
}

// StateMachine represents a PIO state machine
type StateMachine struct {
	chip    *gpiocdev.Chip
	sm      int
	program []uint16
	pins    []int
	mu      sync.Mutex
	pio     *PIO
}

// Config holds the configuration for a state machine
type Config struct {
	ChipNumber string
	SMNumber   int
	Program    []uint16
	Pins       []int
}

// NewPIO creates a new PIO controller
func NewPIO() (*PIO, error) {
	chip, err := gpiocdev.NewChip("gpiochip0")
	if err != nil {
		return nil, fmt.Errorf("failed to open gpiochip0: %v", err)
	}

	// Open /dev/mem for direct memory access
	pio, err := os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		chip.Close()
		return nil, fmt.Errorf("failed to open /dev/mem for PIO: %v", err)
	}

	// Map PIO memory
	mem, err := mapMemory(pio, PIOBaseAddr, PIOMemSize)
	if err != nil {
		pio.Close()
		chip.Close()
		return nil, fmt.Errorf("failed to map PIO memory: %v", err)
	}

	return &PIO{
		chip: chip,
		pio: pio,
		mem: mem,
	}, nil
}

// mapMemory maps a region of physical memory
func mapMemory(f *os.File, addr, size uint32) ([]byte, error) {
	// Map memory with correct size and alignment
	mem, err := unix.Mmap(
		int(f.Fd()),
		int64(addr),
		int(size),
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED,
	)
	if err != nil {
		return nil, fmt.Errorf("mmap failed: %v", err)
	}

	return mem, nil
}

// Close closes the PIO controller
func (p *PIO) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mem != nil {
		if err := unix.Munmap(p.mem); err != nil {
			return fmt.Errorf("munmap failed: %v", err)
		}
		p.mem = nil
	}

	if p.pio != nil {
		p.pio.Close()
		p.pio = nil
	}

	if p.chip != nil {
		p.chip.Close()
		p.chip = nil
	}

	return nil
}

// ConfigurePin configures a GPIO pin for output
func (p *PIO) ConfigurePin(pin int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Configure pin as output
	_, err := p.chip.RequestLine(pin, gpiocdev.AsOutput(0))
	if err != nil {
		return fmt.Errorf("failed to configure pin %d: %v", pin, err)
	}

	return nil
}

// readReg reads a register value
func (p *PIO) readReg(addr uint32) (uint32, error) {
	if p.mem == nil {
		return 0, fmt.Errorf("memory not mapped")
	}

	offset := addr - PIOBaseAddr
	if offset >= uint32(len(p.mem)) {
		return 0, fmt.Errorf("register address out of range: 0x%x", addr)
	}

	// Read 32-bit value from memory
	val := *(*uint32)(unsafe.Pointer(&p.mem[offset]))
	return val, nil
}

// writeReg writes a register value
func (p *PIO) writeReg(addr uint32, val uint32) error {
	if p.mem == nil {
		return fmt.Errorf("memory not mapped")
	}

	offset := addr - PIOBaseAddr
	if offset >= uint32(len(p.mem)) {
		return fmt.Errorf("register address out of range: 0x%x", addr)
	}

	// Write 32-bit value to memory
	*(*uint32)(unsafe.Pointer(&p.mem[offset])) = val
	return nil
}

// NewStateMachine creates a new PIO state machine
func NewStateMachine(cfg Config) (*StateMachine, error) {
	chip, err := gpiocdev.NewChip(cfg.ChipNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to open GPIO chip: %v", err)
	}

	// Create PIO controller
	pio, err := NewPIO()
	if err != nil {
		chip.Close()
		return nil, fmt.Errorf("failed to create PIO controller: %v", err)
	}

	sm := &StateMachine{
		chip:    chip,
		sm:      cfg.SMNumber,
		program: cfg.Program,
		pins:    cfg.Pins,
		pio:     pio,
	}

	if err := sm.init(); err != nil {
		chip.Close()
		pio.Close()
		return nil, err
	}

	return sm, nil
}

// init initializes the state machine
func (sm *StateMachine) init() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Configure pins
	for _, pin := range sm.pins {
		if err := sm.ConfigurePin(pin); err != nil {
			return fmt.Errorf("failed to configure pin %d: %v", pin, err)
		}
	}

	// Load program
	if err := sm.loadProgram(); err != nil {
		return fmt.Errorf("failed to load program: %v", err)
	}

	return nil
}

// ConfigurePin configures a GPIO pin for output
func (sm *StateMachine) ConfigurePin(pin int) error {
	_, err := sm.chip.RequestLine(pin, gpiocdev.AsOutput(0))
	if err != nil {
		return fmt.Errorf("failed to configure pin %d: %v", pin, err)
	}
	return nil
}

// loadProgram loads the PIO program into the state machine
func (sm *StateMachine) loadProgram() error {
	if sm.pio == nil {
		return fmt.Errorf("PIO controller not initialized")
	}

	// Write program to instruction memory
	for i, instr := range sm.program {
		if err := sm.pio.writeReg(PIOBaseAddr+uint32(i*2), uint32(instr)); err != nil {
			return fmt.Errorf("failed to write instruction %d: %v", i, err)
		}
	}

	// Configure state machine
	smOffset := uint32(sm.sm * 0x40)

	// Set clock divider for ~1MHz
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_CLKDIV, 0x1000); err != nil {
		return fmt.Errorf("failed to set clock divider: %v", err)
	}

	// Configure shift control for 32-bit output, shift right
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_SHIFTCTRL, 0x80000000); err != nil {
		return fmt.Errorf("failed to set shift control: %v", err)
	}

	// Configure pins for output
	pinctrl := uint32(0)
	pinctrl |= uint32(sm.pins[0])         // Base pin
	pinctrl |= uint32(len(sm.pins)-1) << 26 // Number of pins - 1
	pinctrl |= uint32(1) << 5             // OUT_EN
	pinctrl |= uint32(1) << 7             // SET_EN
	pinctrl |= uint32(1) << 20            // SIDESET_EN
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_PINCTRL, pinctrl); err != nil {
		return fmt.Errorf("failed to set pin control: %v", err)
	}

	// Set program counter to start
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_ADDR, 0); err != nil {
		return fmt.Errorf("failed to set program counter: %v", err)
	}

	return nil
}

// Start starts the state machine
func (sm *StateMachine) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.pio == nil {
		return fmt.Errorf("PIO controller not initialized")
	}

	smOffset := uint32(sm.sm * 0x40)

	// Set execution control to start the state machine
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_EXECCTRL, 0x1); err != nil {
		return fmt.Errorf("failed to start state machine: %v", err)
	}

	return nil
}

// Stop stops the state machine
func (sm *StateMachine) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.pio == nil {
		return fmt.Errorf("PIO controller not initialized")
	}

	smOffset := uint32(sm.sm * 0x40)

	// Set execution control to stop the state machine
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_EXECCTRL, 0x0); err != nil {
		return fmt.Errorf("failed to stop state machine: %v", err)
	}

	return nil
}

// Put puts data into the state machine's TX FIFO
func (sm *StateMachine) Put(data uint32) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.pio == nil {
		return fmt.Errorf("PIO controller not initialized")
	}

	smOffset := uint32(sm.sm * 0x40)

	// Wait for FIFO space with timeout
	deadline := time.Now().Add(time.Millisecond * 100)
	for {
		fstat, err := sm.pio.readReg(PIOBaseAddr + smOffset + SM0_FSTAT)
		if err != nil {
			return fmt.Errorf("failed to read FIFO status: %v", err)
		}

		if (fstat & 0x1) == 0 {
			// FIFO has space
			break
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for FIFO space")
		}

		time.Sleep(time.Microsecond * 100)
	}

	// Write data to FIFO
	if err := sm.pio.writeReg(PIOBaseAddr+smOffset+SM0_TXF, data); err != nil {
		return fmt.Errorf("failed to write to FIFO: %v", err)
	}

	return nil
}

// Close closes the state machine and releases resources
func (sm *StateMachine) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if err := sm.Stop(); err != nil {
		return err
	}

	if sm.chip != nil {
		sm.chip.Close()
		sm.chip = nil
	}

	if sm.pio != nil {
		sm.pio.Close()
		sm.pio = nil
	}

	return nil
}

// ConfigureHUB75Pins sets up all GPIO pins needed for HUB75
func (p *PIO) ConfigureHUB75Pins() error {
	pins := []int{
		R1_PIN, G1_PIN, B1_PIN,
		R2_PIN, G2_PIN, B2_PIN,
		A_PIN, B_PIN, C_PIN,
		CLK_PIN, LAT_PIN, OE_PIN,
	}

	for _, pin := range pins {
		if err := p.ConfigurePin(pin); err != nil {
			return fmt.Errorf("failed to configure pin %d: %v", pin, err)
		}
	}

	return nil
}

// WriteLEDData writes RGB data for a single row
func (p *PIO) WriteLEDData(rowData []byte, row int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate FIFO address for this state machine
	fifo := 0x200 + uint32(row)*0x10

	// Set row address
	rowAddr := uint32(row & 0x7) // Assuming 8 rows (3 address bits)
	p.writeReg(fifo, rowAddr)

	// Write RGB data
	for i := 0; i < len(rowData); i += 3 {
		r := rowData[i]
		g := rowData[i+1]
		b := rowData[i+2]

		// Pack RGB data
		data := uint32(r)<<16 | uint32(g)<<8 | uint32(b)
		p.writeReg(fifo+4, data)
	}

	// Latch data
	p.writeReg(fifo+0, 0xFF)

	return nil
}

// WriteFIFO writes RGB LED data to the state machine's FIFO
func (p *PIO) WriteFIFO(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate FIFO address for this state machine
	fifo := 0x200 + uint32(data[0])*0x10

	// Convert RGB data to WS2812B bit stream
	for i := 0; i < len(data); i += 3 {
		r := data[i]
		g := data[i+1]
		b := data[i+2]

		// WS2812B expects GRB order
		bits := uint32(g)<<16 | uint32(r)<<8 | uint32(b)

		// Write 24 bits to FIFO
		p.writeReg(fifo+4, bits)
	}

	// Add reset code (zeros)
	p.writeReg(fifo+0, 0)
	p.writeReg(fifo+1, 0)
	p.writeReg(fifo+2, 0)

	return nil
}

// ReadFIFO reads data from the state machine's FIFO
func (p *PIO) ReadFIFO() (uint32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate FIFO address for this state machine
	rxf, err := p.readReg(SM0_RXF)
	if err != nil {
		return 0, err
	}
	fifo := 0x200 + uint32(rxf)*0x10

	// Read 32 bits from FIFO
	val0, err := p.readReg(fifo)
	if err != nil {
		return 0, err
	}
	val1, err := p.readReg(fifo + 4)
	if err != nil {
		return 0, err
	}
	val2, err := p.readReg(fifo + 8)
	if err != nil {
		return 0, err
	}
	val3, err := p.readReg(fifo + 12)
	if err != nil {
		return 0, err
	}

	value := val0 | (val1 << 8) | (val2 << 16) | (val3 << 24)
	return value, nil
}

// IsFIFOFull checks if the FIFO is full
func (p *PIO) IsFIFOFull() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Read FIFO status register
	fstat, err := p.readReg(SM0_FSTAT)
	if err != nil {
		return false, err
	}
	base := 0x100 + uint32(fstat)*0x40

	val0, err := p.readReg(base + SM0_FSTAT)
	if err != nil {
		return false, err
	}
	val1, err := p.readReg(base + SM0_FSTAT + 4)
	if err != nil {
		return false, err
	}
	val2, err := p.readReg(base + SM0_FSTAT + 8)
	if err != nil {
		return false, err
	}
	val3, err := p.readReg(base + SM0_FSTAT + 12)
	if err != nil {
		return false, err
	}

	status := val0 | (val1 << 8) | (val2 << 16) | (val3 << 24)
	return (status & 0x1) != 0, nil
}

// IsFIFOEmpty checks if the FIFO is empty
func (p *PIO) IsFIFOEmpty() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Read FIFO status register
	fstat, err := p.readReg(SM0_FSTAT)
	if err != nil {
		return false, err
	}
	base := 0x100 + uint32(fstat)*0x40

	val0, err := p.readReg(base + SM0_FSTAT)
	if err != nil {
		return false, err
	}
	val1, err := p.readReg(base + SM0_FSTAT + 4)
	if err != nil {
		return false, err
	}
	val2, err := p.readReg(base + SM0_FSTAT + 8)
	if err != nil {
		return false, err
	}
	val3, err := p.readReg(base + SM0_FSTAT + 12)
	if err != nil {
		return false, err
	}

	status := val0 | (val1 << 8) | (val2 << 16) | (val3 << 24)
	return (status & 0x2) != 0, nil
}

// WaitForFIFO waits for the FIFO to be ready
func (p *PIO) WaitForFIFO(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		full, err := p.IsFIFOFull()
		if err != nil {
			return err
		}
		if !full {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for FIFO")
		}
		time.Sleep(time.Microsecond)
	}
} 