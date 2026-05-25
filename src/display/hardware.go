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
	Height = int16(320)
	// 2.0 inch 240x320 ST7789V2 modules usually expose the full GRAM area.
	// If the picture is shifted after wiring is confirmed, try RowOffset 0/80.
	RowOffset    = int16(0)
	ColumnOffset = int16(0)
	SPIFrequency = uint32(40_000_000)

	// GMT020-02-8P 8-pin ST7789V2 wiring.
	// BL -> 3V3, CS -> GPIO10, DC -> GPIO11, RST -> GPIO12,
	// SDA -> GPIO14, SCL -> GPIO13, VCC -> 3V3, GND -> GND.
	SCKPin       = machine.GPIO13
	MOSIPin      = machine.GPIO14
	ResetPin     = machine.GPIO12
	DCPin        = machine.GPIO11
	BacklightPin = machine.NoPin
	CSPin        = machine.GPIO10
)

var (
	panel st7789.Device
	ready bool
)

func Setup() {
	if !Enabled {
		return
	}

	bus := machine.SPI0
	if err := bus.Configure(machine.SPIConfig{
		Frequency: SPIFrequency,
		SCK:       SCKPin,
		SDO:       MOSIPin,
		SDI:       machine.NoPin,
		CS:        machine.NoPin,
		Mode:      machine.Mode0,
	}); err != nil {
		diag.Error("display SPI config failed: " + err.Error())
		return
	}
	panel = st7789.New(bus, ResetPin, DCPin, CSPin, BacklightPin)
	panel.Configure(st7789.Config{
		Width:        Width,
		Height:       Height,
		Rotation:     drivers.Rotation0,
		RowOffset:    RowOffset,
		ColumnOffset: ColumnOffset,
		FrameRate:    st7789.FRAMERATE_111,
	})
	ready = true
	diag.Log("display ready: ST7789V2 240x320 SPI 40MHz")
	RenderSelfTest()
}

func Ready() bool {
	return ready
}
