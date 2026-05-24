package display

import (
	"time"

	"esp32s3-demo/src/diag"
)

var logLines [11]string

func StartDebugger() {
	if !ready {
		return
	}
	go func() {
		for {
			RenderDebugger()
			time.Sleep(1200 * time.Millisecond)
		}
	}()
}

func RenderDebugger() {
	if !ready {
		return
	}

	state, count := diag.Snapshot(logLines[:])
	statusColor := colorYellow
	if state.Mode == "STA" && state.IP != "" {
		statusColor = colorGreen
	}
	if state.LastError != "" {
		statusColor = colorRed
	}

	panel.FillScreen(colorBackground)

	drawWiFiIcon(10, 8, statusColor)
	writeTitle(34, 24, "WiFi", colorText)
	drawStatusDot(224, 14, statusColor)

	ssid := state.SSID
	if ssid == "" {
		ssid = "(none)"
	}
	writeSmall(10, 48, "SSID: "+ssid, colorText)
	writeTiny(10, 64, "MODE "+state.Mode+"  IP "+state.IP, colorMuted)

	drawPanel(8, 78, 224, 150)
	writeSmall(16, 96, "DEBUG LOG", colorBlue)

	first := 0
	if count > len(logLines) {
		count = len(logLines)
	}
	if count > 9 {
		first = count - 9
	}
	y := int16(116)
	for i := first; i < count; i++ {
		line := trimLogPrefix(logLines[i])
		writeTiny(16, y, line, colorText)
		y += 12
	}

	if state.LastError != "" {
		writeTiny(16, 222, "ERR "+state.LastError, colorRed)
	} else if state.CachePresent {
		writeTiny(16, 222, "CACHE OK", colorGreen)
	} else {
		writeTiny(16, 222, "CACHE EMPTY", colorMuted)
	}
}
