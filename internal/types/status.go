package types

import (
	"time"
)

// MachineState represents the current state of the FluidNC machine
type MachineState string

const (
	// Possible machine states
	StateIdle      MachineState = "Idle"
	StateRun       MachineState = "Run"
	StateHold      MachineState = "Hold"
	StateJog       MachineState = "Jog"
	StateAlarm     MachineState = "Alarm"
	StateDoor      MachineState = "Door"
	StateCheck     MachineState = "Check"
	StateHome      MachineState = "Home"
	StateSleep     MachineState = "Sleep"
	StateUnknown   MachineState = "Unknown"
)

// Coordinates represents the X, Y, Z coordinates of the machine
type Coordinates struct {
	X float64
	Y float64
	Z float64
}

// MachineStatus represents the complete status of the FluidNC machine
type MachineStatus struct {
	State       MachineState
	Coordinates Coordinates
	FeedRate    float64
	SpindleSpeed float64
	BufferState  int
	LineNumber   int
	LastUpdated  time.Time
}

// DisplayData represents the data to be displayed on the LED matrix
type DisplayData struct {
	MachineStatus MachineStatus
	IPAddress     string
	Connected     bool
	LastUpdated   time.Time
}

// MatrixConfig represents the configuration for the LED matrix
type MatrixConfig struct {
	Pinout            string
	NumAddressLines   int
	NumPlanes         int
	Orientation       string
	Brightness        float64
	NumTemporalPlanes int
}

// DisplayConfig represents the configuration for the display
type DisplayConfig struct {
	Width           int
	Height          int
	Brightness      int
	UpdateInterval  float64
}

// FluidNCConfig represents the configuration for the FluidNC connection
type FluidNCConfig struct {
	Host              string
	Port              int
	ReconnectInterval int
	StatusInterval    float64
}

// DiscoveryConfig represents the configuration for the FluidNC discovery
type DiscoveryConfig struct {
	ScanInterval int
	Timeout      int
} 