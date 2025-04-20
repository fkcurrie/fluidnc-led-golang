package mmap

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// MemoryMap represents a memory mapped region
type MemoryMap struct {
	addr   uintptr
	size   uintptr
	region []byte
}

// NewMemoryMap creates a new memory mapped region
func NewMemoryMap(addr, size uintptr) (*MemoryMap, error) {
	// Open /dev/mem
	f, err := os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/mem: %v", err)
	}
	defer f.Close()

	// Map the memory region
	region, err := syscall.Mmap(
		int(f.Fd()),
		int64(addr),
		int(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to mmap: %v", err)
	}

	return &MemoryMap{
		addr:   addr,
		size:   size,
		region: region,
	}, nil
}

// Close unmaps the memory region
func (m *MemoryMap) Close() error {
	return syscall.Munmap(m.region)
}

// Region returns the mapped memory region
func (m *MemoryMap) Region() []byte {
	return m.region
}

// Read32 reads a 32-bit value from the memory region
func (m *MemoryMap) Read32(offset uintptr) uint32 {
	return *(*uint32)(unsafe.Pointer(&m.region[offset]))
}

// Write32 writes a 32-bit value to the memory region
func (m *MemoryMap) Write32(offset uintptr, value uint32) {
	*(*uint32)(unsafe.Pointer(&m.region[offset])) = value
}

// Read16 reads a 16-bit value from the memory region
func (m *MemoryMap) Read16(offset uintptr) uint16 {
	return *(*uint16)(unsafe.Pointer(&m.region[offset]))
}

// Write16 writes a 16-bit value to the memory region
func (m *MemoryMap) Write16(offset uintptr, value uint16) {
	*(*uint16)(unsafe.Pointer(&m.region[offset])) = value
}

// Read8 reads an 8-bit value from the memory region
func (m *MemoryMap) Read8(offset uintptr) uint8 {
	return m.region[offset]
}

// Write8 writes an 8-bit value to the memory region
func (m *MemoryMap) Write8(offset uintptr, value uint8) {
	m.region[offset] = value
}

// WriteBytes writes a byte slice to the memory region
func (m *MemoryMap) WriteBytes(offset uintptr, data []byte) {
	copy(m.region[offset:], data)
}

// ReadBytes reads a byte slice from the memory region
func (m *MemoryMap) ReadBytes(offset uintptr, size int) []byte {
	return m.region[offset : offset+uintptr(size)]
} 