package rgbmatrix

import "C"
import (
	"github.com/jmaitrehenry/go-rpi-rgb-led-matrix/emulator"
	"github.com/jmaitrehenry/go-rpi-rgb-led-matrix/julien"
	"github.com/jmaitrehenry/go-rpi-rgb-led-matrix/terminal"
	"image/color"
	"os"
	"runtime"
	"strings"
)

// DefaultConfig default WS281x configuration
var DefaultConfig = HardwareConfig{
	Rows:              32,
	Cols:              32,
	ChainLength:       1,
	Parallel:          1,
	PWMBits:           11,
	PWMLSBNanoseconds: 130,
	Brightness:        100,
	ScanMode:          Progressive,
}

// HardwareConfig rgb-led-matrix configuration
type HardwareConfig struct {
	// Rows the number of rows supported by the display, so 32 or 16.
	Rows int
	// Cols the number of columns supported by the display, so 32 or 64 .
	Cols int
	// ChainLengthis the number of displays daisy-chained together
	// (output of one connected to input of next).
	ChainLength int
	// Parallel is the number of parallel chains connected to the Pi; in old Pis
	// with 26 GPIO pins, that is 1, in newer Pis with 40 interfaces pins, that
	// can also be 2 or 3. The effective number of pixels in vertical direction is
	// then thus rows * parallel.
	Parallel int
	// Set PWM bits used for output. Default is 11, but if you only deal with
	// limited comic-colors, 1 might be sufficient. Lower require less CPU and
	// increases refresh-rate.
	PWMBits int
	// Change the base time-unit for the on-time in the lowest significant bit in
	// nanoseconds.  Higher numbers provide better quality (more accurate color,
	// less ghosting), but have a negative impact on the frame rate.
	PWMLSBNanoseconds int // the DMA channel to use
	// Brightness is the initial brightness of the panel in percent. Valid range
	// is 1..100
	Brightness int
	// ScanMode progressive or interlaced
	ScanMode ScanMode // strip color layout
	// A string describing a sequence of pixel mappers that should be applied
	// to this matrix. A semicolon-separated list of pixel-mappers with optional
	// parameter.
	PixelMapperConfig string
	// Disable the PWM hardware subsystem to create pulses. Typically, you don't
	// want to disable hardware pulsing, this is mostly for debugging and figuring
	// out if there is interference with the sound system.
	// This won't do anything if output enable is not connected to GPIO 18 in
	// non-standard wirings.
	DisableHardwarePulsing bool

	ShowRefreshRate bool
	InverseColors   bool

	// Name of GPIO mapping used
	HardwareMapping string
}

func (c *HardwareConfig) geometry() (width, height int) {
	return c.Cols * c.ChainLength, c.Rows * c.Parallel
}

type ScanMode int8

const (
	Progressive ScanMode = 0
	Interlaced  ScanMode = 1
)

// RGBLedMatrix matrix representation for ws281x
type RGBLedMatrix struct {
	Config *HardwareConfig

	height int
	width  int
	matrix *C.struct_RGBLedMatrix
	buffer *C.struct_LedCanvas
	leds   []uint32_t
}

const MatrixEmulatorENV = "MATRIX_EMULATOR"
const TerminalMatrixEmulatorENV = "MATRIX_TERMINAL_EMULATOR"

func isJulienEmulator() bool {
	return runtime.GOOS == "darwin"
}

func isMatrixEmulator() bool {
	if os.Getenv(MatrixEmulatorENV) == "1" {
		return true
	}

	return false
}

func isTerminalMatrixEmulator() bool {
	if os.Getenv(TerminalMatrixEmulatorENV) == "1" {
		return true
	}
	return false
}

func buildJulienMatrix(config *HardwareConfig) Matrix {
	w, h := config.geometry()
	matrix := julien.GenerateEmpty(h, w)
	return &matrix
}

func buildMatrixEmulator(config *HardwareConfig) Matrix {
	w, h := config.geometry()
	return emulator.NewEmulator(w, h, emulator.DefaultPixelPitch, true)
}

func buildTerminalMatrixEmulator(config *HardwareConfig) Matrix {
	w, h := config.geometry()
	if strings.Contains(config.PixelMapperConfig, "U-mapper") {
		w /= 2
		h *= 2
	}
	return terminal.NewTerminal(w, h, true)
}

// Initialize library, must be called once before other functions are
// called.
func (c *RGBLedMatrix) Initialize() error {
	return nil
}

// Geometry returns the width and the height of the matrix
func (c *RGBLedMatrix) Geometry() (width, height int) {
	return c.width, c.height
}

func colorToUint32(c color.Color) uint32 {
	if c == nil {
		return 0
	}

	// A color's RGBA method returns values in the range [0, 65535]
	red, green, blue, _ := c.RGBA()
	return (red>>8)<<16 | (green>>8)<<8 | blue>>8
}

func uint32ToColor(u uint32_t) color.Color {
	return color.RGBA{
		uint8(u>>16) & 255,
		uint8(u>>8) & 255,
		uint8(u>>0) & 255,
		0,
	}
}
