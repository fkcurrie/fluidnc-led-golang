# GitHub Setup and Issue Creation Commands

This guide provides commands to initialize a GitHub repository, commit the code, create a release, and set up issues for the FluidNC LED Matrix project.

## Repository Setup

1. Initialize Git repository (if not already done):
```bash
git init
```

2. Add all files:
```bash
git add .
```

3. Make the initial commit:
```bash
git commit -m "Initial commit: Basic HUB75 LED matrix driver for Raspberry Pi 5"
```

4. Create a new repository on GitHub through the web interface:
   - Go to: https://github.com/new
   - Name: `fluidnc-led-golang`
   - Description: `Go implementation of an LED matrix display for FluidNC on Raspberry Pi 5`
   - Choose public repository
   - Click "Create repository"

5. Add the remote and push:
```bash
git remote add origin https://github.com/yourusername/fluidnc-led-golang.git
git branch -M main
git push -u origin main
```

## Creating a Release

1. Create a tag for the release:
```bash
git tag -a v0.1.0 -m "Early release with basic HUB75 functionality"
git push origin v0.1.0
```

2. On GitHub, go to the "Releases" section:
   - Click on "Draft a new release"
   - Choose the v0.1.0 tag
   - Title: "v0.1.0 - Early Release"
   - Description: Copy the details below
   - Check "This is a pre-release"
   - Click "Publish release"

**Release Description:**
```
First early release of the FluidNC LED Matrix driver.

## What works
- Direct GPIO control of HUB75 LED matrix via Adafruit RGB Matrix Bonnet
- Text scrolling capability with 5x7 pixel font
- Test patterns for display verification (red, green, blue, checkerboard)
- Configuration management

## Known issues
- Text scrolling shows flickering under certain conditions
- Text positioning may appear duplicated or incorrect
- Limited to single-bit color depth
- No connection to FluidNC status API yet

## Next steps
See open issues for planned improvements.
```

## Creating Issues

Create the following issues:

### Issue 1: Fix text scrolling issues
```
Title: Fix text scrolling issues (flickering and positioning)

Description:
The current text scrolling implementation has two main issues:
1. Text sometimes appears to flicker during scrolling
2. Text may appear duplicated or at incorrect positions

Tasks:
- [ ] Review the rendering logic in the text scrolling code
- [ ] Fix the vertical positioning to ensure text is displayed in a single line
- [ ] Eliminate any flickering effects during animation
- [ ] Ensure correct double-buffering to prevent partial updates
- [ ] Add proper bounds checking for panel size
- [ ] Test with different text content and lengths
```

### Issue 2: Implement FluidNC WebSocket connection
```
Title: Implement FluidNC WebSocket client for real-time status

Description:
Create a WebSocket client to connect to the FluidNC API and retrieve real-time machine status.

Tasks:
- [ ] Implement WebSocket client using Gorilla WebSocket library
- [ ] Connect to FluidNC's WebSocket endpoint
- [ ] Parse incoming JSON status messages
- [ ] Create data structures to hold machine status information
- [ ] Implement reconnection logic for network interruptions
- [ ] Add configuration options for FluidNC server address/port
```

### Issue 3: Create status display layouts
```
Title: Design and implement status display layouts

Description:
Create configurable display layouts to show FluidNC machine status on the LED matrix.

Tasks:
- [ ] Design layout for machine position (X, Y, Z coordinates)
- [ ] Design layout for machine state (Idle, Run, Hold, etc.)
- [ ] Add layout for tool information (spindle speed, tool number)
- [ ] Create layout for job progress
- [ ] Implement layout manager to switch between different screens
- [ ] Add configuration options for customizing layouts
```

### Issue 4: Add multi-bit color depth support
```
Title: Implement multi-bit color depth for better display quality

Description:
The current implementation uses single-bit color depth (on/off). Implement multi-bit color depth for better visual quality.

Tasks:
- [ ] Research bit-angle modulation (BAM) or pulse-width modulation (PWM) techniques for LED displays
- [ ] Modify the driver to support multiple brightness levels for each color channel
- [ ] Implement efficient time-multiplexing for different bit depths
- [ ] Add color conversion from RGB values to LED brightness levels
- [ ] Update rendering functions to support grayscale or full color
```

### Issue 5: Add SVG rendering support
```
Title: Add SVG rendering support

Description:
Implement the ability to display SVG graphics on the LED matrix.

Tasks:
- [ ] Add SVG parsing and rendering capabilities
- [ ] Implement vector to raster conversion at LED matrix resolution
- [ ] Create a simple API for loading and displaying SVGs
- [ ] Add support for basic animations through SVG manipulation
- [ ] Test with a set of example SVG files of different complexity
```

## Additional Resources

- [GitHub CLI Documentation](https://cli.github.com/manual/)
- [Git Cheat Sheet](https://education.github.com/git-cheat-sheet-education.pdf)
- [GitHub Issues Documentation](https://docs.github.com/en/issues)
- [GitHub Actions for Automated Testing](https://docs.github.com/en/actions)
