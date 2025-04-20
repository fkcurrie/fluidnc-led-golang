# FluidNC LED Screen Monitor

A Go implementation of an LED matrix display monitor for FluidNC, designed for Raspberry Pi 5 using the Adafruit RGB Matrix Bonnet.

## Status: Early Release

This is an early release with basic functionality. The code successfully drives a 32x64 LED matrix via the Adafruit RGB Matrix Bonnet, but there are several known issues and improvements planned (see issues on GitHub).

## Project Structure

```
.
├── cmd/
│   ├── hub75-gpio/    # Main program using direct GPIO for LED matrix control
│   └── gpio-test/     # Simple test program for GPIO access
├── pkg/
│   └── gpio/          # GPIO control package
└── internal/
    ├── config/        # Configuration management
    └── types/         # Shared type definitions
```

## Features

- Direct GPIO control of HUB75 LED matrix via Adafruit RGB Matrix Bonnet
- Text scrolling capability with 5x7 pixel font
- Test patterns for display verification
- Configuration management via JSON
- Graceful shutdown handling

## Installation

1. Install the Adafruit RGB Matrix Bonnet on your Raspberry Pi 5
2. Clone the repository:
```bash
git clone https://github.com/yourusername/fluidnc-led-golang.git
cd fluidnc-led-golang
```

3. Build the program:
```bash
go build -o fluidnc-led ./cmd/hub75-gpio
```

## Usage

Run the program with sudo (required for GPIO access):
```bash
sudo ./fluidnc-led
```

### Command Line Flags

- `-text="YOUR TEXT"`: Sets the text to scroll across the display
- `-scroll`: Shows scrolling text instead of test patterns

## Panel Configuration

This implementation is designed for a 32x64 LED matrix panel using the HUB75 interface via the Adafruit RGB Matrix Bonnet:

- Panel height: 32 pixels
- Panel width: 64 pixels
- Addressable rows: 16 (each row controls both upper and lower half)

The HUB75 LED matrix uses the following BCM GPIO pins as specified by the Adafruit RGB Matrix Bonnet documentation:

- R1Pin (GPIO 5): Red data for upper half
- G1Pin (GPIO 13): Green data for upper half
- B1Pin (GPIO 6): Blue data for upper half
- R2Pin (GPIO 12): Red data for lower half
- G2Pin (GPIO 16): Green data for lower half
- B2Pin (GPIO 23): Blue data for lower half
- CLKPin (GPIO 17): Clock signal
- OEPin (GPIO 4): Output enable
- LAPin (GPIO 21): Latch signal
- ABPin (GPIO 22): Address bit A
- BCPin (GPIO 26): Address bit B
- CCPin (GPIO 27): Address bit C
- DPin (GPIO 20): Address bit D
- EPin (GPIO 24): Address bit E (only needed for 64-pixel high displays)

## Known Issues

- Text scrolling shows flickering under certain conditions
- Sometimes text appears duplicated or at incorrect positions
- Single-bit color depth limits display capabilities
- No connection to FluidNC status API yet

## Notes

- The program requires root access to control GPIO pins
- Our implementation successfully uses direct GPIO control via the go-gpiocdev library
- Each GPIO pin is individually controlled through the Linux GPIO character device
- Binary data (0 or 1) is used for each color channel (single bit color depth)
- The HUB75 protocol is implemented in software with precise timing
- For better performance, consider researching PWM control for multi-bit color depth
- If GPIO pins show as "device or resource busy", you may need to check what's using them:
  ```bash
  sudo cat /sys/kernel/debug/gpio
  ```
- Ensure your user is in the gpio group and proper udev rules are set up:
  ```bash
  # /etc/udev/rules.d/99-com.rules
  SUBSYSTEM=="gpio", KERNEL=="gpiochip[0-9]*", GROUP="gpio", MODE="0660"
  SUBSYSTEM=="*-pio", GROUP="gpio", MODE="0660"
  ```

### Working Configuration

We've confirmed that our HUB75 LED matrix driver works with the following approach:

1. Using the go-gpiocdev library for direct GPIO access
2. Using the correct chip name "gpiochip0" for the Raspberry Pi 5 GPIO pins
3. Accessing pins with their BCM GPIO numbers (e.g., 5 for GPIO5)
4. Implementing the HUB75 protocol in software with the following sequence:
   - Set row address bits
   - Disable output
   - For each pixel:
     - Set RGB data pins
     - Pulse the clock
   - Latch the data
   - Enable output
5. Adding small delays between operations to ensure proper timing

### Troubleshooting

If you encounter "device or resource busy" errors:

1. Check which processes are using GPIO pins:
   ```bash
   sudo cat /sys/kernel/debug/gpio
   ```

2. The pins might be used by the kernel or other drivers. You may need to disable those drivers
   or configure the device tree to release those pins.

3. Use the `pinctrl` tool to check the current GPIO configuration:
   ```bash
   pinctrl get | grep -i gpio
   ```

4. Try our simple GPIO test program to verify GPIO access works:
   ```bash
   sudo go run cmd/gpio-test/main.go
   ```

## Development Roadmap

Future development plans include:
- Fix text scrolling issues (flickering, positioning)
- Implement FluidNC WebSocket client for real-time status
- Add multi-bit color depth support for better color rendering
- Create configurable display layouts for showing machine position, status, etc.
- Add SVG rendering support
- Develop efficient double-buffering for smoother animations

## Building from Source

```bash
go build -o fluidnc-led ./cmd/hub75-gpio
```

## License

MIT License 