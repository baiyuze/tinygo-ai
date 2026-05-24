package pinprobe

import (
	"machine"
	"time"
)

type probePin struct {
	name string
	pin  machine.Pin
}

var pins = [...]probePin{
	{name: "GPIO12 SCK", pin: machine.GPIO12},
	{name: "GPIO11 SDA", pin: machine.GPIO11},
	{name: "GPIO7 DC/RST-A", pin: machine.GPIO7},
	{name: "GPIO6 DC/RST-B", pin: machine.GPIO6},
	{name: "GPIO36 original-SCK", pin: machine.GPIO36},
	{name: "GPIO35 original-SDA", pin: machine.GPIO35},
}

func Run() {
	println("[pinprobe] start")
	for _, p := range pins {
		p.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		p.pin.Low()
	}

	for {
		for _, p := range pins {
			println("[pinprobe] HIGH", p.name)
			p.pin.High()
			time.Sleep(2 * time.Second)
			println("[pinprobe] LOW", p.name)
			p.pin.Low()
			time.Sleep(800 * time.Millisecond)
		}
	}
}
