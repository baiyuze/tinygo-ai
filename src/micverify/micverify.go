package micverify

import (
	"device/esp"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"time"

	"esp32s3-demo/src/diag"
	"esp32s3-demo/src/display"
)

const (
	Enabled = false

	SCKPin = machine.GPIO6
	WSPin  = machine.GPIO5
	SDPin  = machine.GPIO4

	sampleCount = 192
)

func Run() {
	diag.Log("mic verify mode")
	diag.Log("INMP441 SCK=GPIO6 WS=GPIO5 SD=GPIO4 LR=GND")

	SCKPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	WSPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SDPin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	sckSet, sckSetMask := SCKPin.PortMaskSet()
	sckClear, sckClearMask := SCKPin.PortMaskClear()
	wsSet, wsSetMask := WSPin.PortMaskSet()
	wsClear, wsClearMask := WSPin.PortMaskClear()
	sdMask := uint32(1 << uint32(SDPin))

	// INMP441 L/R is tied to GND, so the microphone should drive the left
	// slot. Philips I2S uses one BCLK delay after each WS transition.
	volatile.StoreUint32(wsClear, wsClearMask)
	volatile.StoreUint32(sckClear, sckClearMask)

	for {
		leftPeak, rightPeak, minSample, maxSample, toggles := sampleBlock(
			sckSet, sckSetMask,
			sckClear, sckClearMask,
			wsSet, wsSetMask,
			wsClear, wsClearMask,
			sdMask,
		)
		level := levelFromPeak(leftPeak)
		display.RenderMicVerify(level, leftPeak, rightPeak, minSample, maxSample, toggles)
		diag.Log("mic left", diag.Int(int(leftPeak)), "right", diag.Int(int(rightPeak)), "level", diag.Int(level), "toggles", diag.Int(toggles))
		time.Sleep(120 * time.Millisecond)
	}
}

func sampleBlock(sckSet *uint32, sckSetMask uint32, sckClear *uint32, sckClearMask uint32, wsSet *uint32, wsSetMask uint32, wsClear *uint32, wsClearMask uint32, sdMask uint32) (uint32, uint32, int32, int32, int) {
	var leftSamples [sampleCount]int32
	var rightSamples [sampleCount]int32
	var leftSum int64
	var rightSum int64
	var toggles int
	var lastHigh bool

	state := interrupt.Disable()
	for i := 0; i < sampleCount; i++ {
		left, right, sampleToggles, high := readFrame(
			sckSet, sckSetMask,
			sckClear, sckClearMask,
			wsSet, wsSetMask,
			wsClear, wsClearMask,
			sdMask,
			lastHigh,
		)
		lastHigh = high
		toggles += sampleToggles
		leftSamples[i] = left
		rightSamples[i] = right
		leftSum += int64(left)
		rightSum += int64(right)
	}
	interrupt.Restore(state)

	leftPeak, leftMin, leftMax := peakAroundMean(leftSamples[:], int32(leftSum/sampleCount))
	rightPeak, _, _ := peakAroundMean(rightSamples[:], int32(rightSum/sampleCount))
	return leftPeak, rightPeak, leftMin, leftMax, toggles
}

func peakAroundMean(samples []int32, avg int32) (uint32, int32, int32) {
	minSample := samples[0]
	maxSample := samples[0]
	var peak uint32
	for _, sample := range samples {
		if sample < minSample {
			minSample = sample
		}
		if sample > maxSample {
			maxSample = sample
		}
		delta := sample - avg
		if delta < 0 {
			delta = -delta
		}
		if uint32(delta) > peak {
			peak = uint32(delta)
		}
	}
	return peak, minSample, maxSample
}

func readFrame(sckSet *uint32, sckSetMask uint32, sckClear *uint32, sckClearMask uint32, wsSet *uint32, wsSetMask uint32, wsClear *uint32, wsClearMask uint32, sdMask uint32, lastHigh bool) (int32, int32, int, bool) {
	volatile.StoreUint32(wsClear, wsClearMask)
	toggles := 0
	high := lastHigh
	left, leftToggles, high := readSlot24(sckSet, sckSetMask, sckClear, sckClearMask, sdMask, high)
	toggles += leftToggles

	volatile.StoreUint32(wsSet, wsSetMask)
	right, rightToggles, high := readSlot24(sckSet, sckSetMask, sckClear, sckClearMask, sdMask, high)
	toggles += rightToggles
	return left, right, toggles, high
}

func readSlot24(sckSet *uint32, sckSetMask uint32, sckClear *uint32, sckClearMask uint32, sdMask uint32, lastHigh bool) (int32, int, bool) {
	var word uint32
	toggles := 0
	high := lastHigh

	toggles, high = pulseDiscard(sckSet, sckSetMask, sckClear, sckClearMask, sdMask, high, toggles)
	for bit := 0; bit < 24; bit++ {
		var nowHigh bool
		nowHigh, high, toggles = pulseSample(sckSet, sckSetMask, sckClear, sckClearMask, sdMask, high, toggles)
		word <<= 1
		if nowHigh {
			word |= 1
		}
	}
	for bit := 0; bit < 7; bit++ {
		toggles, high = pulseDiscard(sckSet, sckSetMask, sckClear, sckClearMask, sdMask, high, toggles)
	}
	return signExtend24(word), toggles, high
}

func pulseDiscard(sckSet *uint32, sckSetMask uint32, sckClear *uint32, sckClearMask uint32, sdMask uint32, lastHigh bool, toggles int) (int, bool) {
	_, high, toggles := pulseSample(sckSet, sckSetMask, sckClear, sckClearMask, sdMask, lastHigh, toggles)
	return toggles, high
}

func pulseSample(sckSet *uint32, sckSetMask uint32, sckClear *uint32, sckClearMask uint32, sdMask uint32, lastHigh bool, toggles int) (bool, bool, int) {
	volatile.StoreUint32(sckSet, sckSetMask)
	nowHigh := esp.GPIO.IN.Get()&sdMask != 0
	if nowHigh != lastHigh {
		toggles++
	}
	volatile.StoreUint32(sckClear, sckClearMask)
	return nowHigh, nowHigh, toggles
}

func signExtend24(raw uint32) int32 {
	if raw&0x800000 != 0 {
		raw |= 0xff000000
	}
	return int32(raw)
}

func levelFromPeak(peak uint32) int {
	level := int(peak >> 14)
	if level > 100 {
		return 100
	}
	return level
}
