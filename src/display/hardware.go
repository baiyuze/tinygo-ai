package display

import (
	"machine"

	"esp32s3-demo/src/diag"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
)

const (
	Enabled = true

	Width  = int16(240)
	Height = int16(240)
	// Many 1.3 inch 240x240 ST7789 modules map the visible area at RAM row 80.
	// If the screen lights but content is shifted/missing, this is the first
	// value to try as 0.
	RowOffset    = int16(80)
	ColumnOffset = int16(0)

	// GMT130-V1.0 7-pin ST7789 wiring.
	// VCC -> 3V3, GND -> GND, SCK/SCL -> GPIO14, SDA -> GPIO13,
	// RST -> GPIO6, DC -> GPIO7, BLK -> 3V3. CS is not present.
	SCKPin       = machine.GPIO14
	MOSIPin      = machine.GPIO13
	ResetPin     = machine.GPIO6
	DCPin        = machine.GPIO7
	BacklightPin = machine.NoPin
	CSPin        = machine.NoPin
)

var (
	panel st7789.Device
	ready bool
)

func Setup() {
	if !Enabled {
		return
	}

	bus := newSoftSPI(SCKPin, MOSIPin)
	panel = st7789.New(bus, ResetPin, DCPin, CSPin, BacklightPin)
	panel.Configure(st7789.Config{
		Width:        Width,
		Height:       Height,
		Rotation:     drivers.Rotation0,
		RowOffset:    RowOffset,
		ColumnOffset: ColumnOffset,
	})
	ready = true
	diag.Log("display ready: ST7789 240x240")
	RenderSelfTest()
}

func Ready() bool {
	return ready
}
