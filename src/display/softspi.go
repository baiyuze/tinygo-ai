package display

import (
	"machine"
)

type softSPI struct {
	sck  machine.Pin
	mosi machine.Pin
}

func newSoftSPI(sck, mosi machine.Pin) *softSPI {
	sck.Configure(machine.PinConfig{Mode: machine.PinOutput})
	mosi.Configure(machine.PinConfig{Mode: machine.PinOutput})
	sck.Low()
	mosi.Low()
	return &softSPI{sck: sck, mosi: mosi}
}

func (s *softSPI) Tx(w, r []byte) error {
	if w == nil {
		for range r {
			_, _ = s.Transfer(0)
		}
		return nil
	}
	for _, b := range w {
		_, _ = s.Transfer(b)
	}
	return nil
}

func (s *softSPI) Transfer(b byte) (byte, error) {
	for mask := byte(0x80); mask != 0; mask >>= 1 {
		if b&mask != 0 {
			s.mosi.High()
		} else {
			s.mosi.Low()
		}
		s.sck.High()
		s.sck.Low()
	}
	return 0, nil
}
