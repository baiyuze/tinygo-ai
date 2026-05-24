package rgb

import (
	"device"
	"image/color"
	"time"
	"unsafe"

	"machine"
	"runtime/interrupt"
)

var ready bool

func Setup() {
	pin := machine.GPIO48
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ready = true
	Set(color.RGBA{})
}

func SetupMode() {
	Set(color.RGBA{B: 24})
}

func WiFiConnected() {
	Set(color.RGBA{G: 24})
}

func Error() {
	Set(color.RGBA{R: 20})
}

func Set(c color.RGBA) {
	if !ready {
		return
	}
	pin := machine.GPIO48
	writeWS2812Byte(pin, c.G)
	writeWS2812Byte(pin, c.R)
	writeWS2812Byte(pin, c.B)
	time.Sleep(80 * time.Microsecond)
}

func writeWS2812Byte(pin machine.Pin, c byte) {
	portSet, maskSet := pin.PortMaskSet()
	portClear, maskClear := pin.PortMaskClear()
	mask := interrupt.Disable()
	device.AsmFull(`
	1:
		s32i  {maskSet}, {portSet}, 0
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop
		slli  {value}, {value}, 1
		bbsi  {value}, 8, 2f
		s32i  {maskClear}, {portClear}, 0
	2:
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop
		s32i  {maskClear}, {portClear}, 0
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop; nop; nop; nop; nop; nop; nop; nop; nop
		nop; nop
		addi  {i}, {i}, -1
		bnez {i}, 1b
		movi.n {i}, 8
		slli  {value}, {value}, 8
	`, map[string]interface{}{
		"value":     uint32(c),
		"i":         8,
		"maskSet":   maskSet,
		"portSet":   uintptr(unsafe.Pointer(portSet)),
		"maskClear": maskClear,
		"portClear": uintptr(unsafe.Pointer(portClear)),
	})
	interrupt.Restore(mask)
}
