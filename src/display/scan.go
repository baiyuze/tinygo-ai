package display

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
)

type scanMapping struct {
	name      string
	sck       machine.Pin
	mosi      machine.Pin
	reset     machine.Pin
	dc        machine.Pin
	rowOffset int16
}

var scanMappings = [...]scanMapping{
	{name: "A sck12 sda11 rst6 dc7 row80", sck: machine.GPIO12, mosi: machine.GPIO11, reset: machine.GPIO6, dc: machine.GPIO7, rowOffset: 80},
	{name: "B sck12 sda11 rst6 dc7 row0", sck: machine.GPIO12, mosi: machine.GPIO11, reset: machine.GPIO6, dc: machine.GPIO7, rowOffset: 0},
	{name: "C sck11 sda12 rst6 dc7 row80", sck: machine.GPIO11, mosi: machine.GPIO12, reset: machine.GPIO6, dc: machine.GPIO7, rowOffset: 80},
	{name: "D sck11 sda12 rst6 dc7 row0", sck: machine.GPIO11, mosi: machine.GPIO12, reset: machine.GPIO6, dc: machine.GPIO7, rowOffset: 0},
	{name: "E sck12 sda11 rst7 dc6 row80", sck: machine.GPIO12, mosi: machine.GPIO11, reset: machine.GPIO7, dc: machine.GPIO6, rowOffset: 80},
	{name: "F sck12 sda11 rst7 dc6 row0", sck: machine.GPIO12, mosi: machine.GPIO11, reset: machine.GPIO7, dc: machine.GPIO6, rowOffset: 0},
	{name: "G sck36 sda35 rst6 dc7 row80", sck: machine.GPIO36, mosi: machine.GPIO35, reset: machine.GPIO6, dc: machine.GPIO7, rowOffset: 80},
	{name: "H sck36 sda35 rst6 dc7 row0", sck: machine.GPIO36, mosi: machine.GPIO35, reset: machine.GPIO6, dc: machine.GPIO7, rowOffset: 0},
}

func RunScan() {
	println("[display-scan] start")
	for {
		for _, m := range scanMappings {
			println("[display-scan] try", m.name)
			bus := newSoftSPI(m.sck, m.mosi)
			dev := st7789.New(bus, m.reset, m.dc, machine.NoPin, machine.NoPin)
			dev.Configure(st7789.Config{
				Width:     Width,
				Height:    Height,
				Rotation:  drivers.Rotation0,
				RowOffset: m.rowOffset,
			})
			dev.FillScreen(colorRed)
			time.Sleep(700 * time.Millisecond)
			dev.FillScreen(colorGreen)
			time.Sleep(700 * time.Millisecond)
			dev.FillScreen(colorBlue)
			time.Sleep(700 * time.Millisecond)
			dev.FillScreen(colorText)
			time.Sleep(1600 * time.Millisecond)
		}
	}
}
