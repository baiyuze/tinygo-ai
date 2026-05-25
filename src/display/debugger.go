package display

import (
	"image/color"
	"time"

	"esp32s3-demo/src/diag"
)

var (
	logLines [11]string

	debuggerScreenReady bool
	lastTopRegionKey    string
	lastFooterRegionKey string
	lastLogLineKeys     [debugLogLineCount]string
)

const debugLogLineCount = 8

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

	status, statusColor := debuggerStatus(state)
	topKey := debuggerTopKey(state, status)
	footerKey := debuggerFooterKey(state)
	logKeys := debuggerLogKeys(count)

	if !debuggerScreenReady {
		panel.FillScreen(colorBackground)
		renderDebuggerStatic()
		debuggerScreenReady = true
	}
	if topKey != lastTopRegionKey {
		renderDebuggerTop(state, statusColor)
		lastTopRegionKey = topKey
	}
	for i := range logKeys {
		if logKeys[i] != lastLogLineKeys[i] {
			renderDebuggerLogLine(i, logKeys[i])
			lastLogLineKeys[i] = logKeys[i]
		}
	}
	if footerKey != lastFooterRegionKey {
		renderDebuggerFooter(state)
		lastFooterRegionKey = footerKey
	}
}

func debuggerStatus(state diag.State) (int, color.RGBA) {
	if state.LastError != "" {
		return 2, colorRed
	}
	if state.Mode == "STA" && state.IP != "" {
		return 1, colorGreen
	}
	return 0, colorYellow
}

func debuggerTopKey(state diag.State, status int) string {
	ssid := state.SSID
	if ssid == "" {
		ssid = "(none)"
	}
	return diag.Int(status) + "|" + ssid + "|" + state.Mode + "|" + state.IP
}

func debuggerFooterKey(state diag.State) string {
	return state.LastError + "|" + boolKey(state.CachePresent)
}

func boolKey(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func debuggerLogKeys(count int) [debugLogLineCount]string {
	var keys [debugLogLineCount]string
	if count > len(logLines) {
		count = len(logLines)
	}
	first := 0
	if count > debugLogLineCount {
		first = count - debugLogLineCount
	}
	for i := first; i < count; i++ {
		keys[i-first] = trimLogPrefix(logLines[i])
	}
	return keys
}

func renderDebuggerStatic() {
	drawPanel(8, 78, 224, 150)
	writeLargeUTF8(16, 96, "日志", colorBlue)
}

func renderDebuggerTop(state diag.State, statusColor color.RGBA) {
	panel.FillRectangle(0, 0, Width, 72, colorBackground)
	drawWiFiIcon(10, 8, statusColor)
	writeLargeUTF8(34, 24, "WiFi", colorText)
	drawStatusDot(224, 14, statusColor)

	ssid := state.SSID
	if ssid == "" {
		ssid = "(none)"
	}
	writeLargeUTF8(10, 48, ssid, colorText)
	writeSmall(10, 64, "MODE "+state.Mode+"  IP "+state.IP, colorMuted)
}

func renderDebuggerLogLine(index int, line string) {
	y := int16(116 + index*14)
	panel.FillRectangle(12, y-11, 216, 13, colorPanel)
	if line != "" {
		writeLargeUTF8(16, y, line, colorText)
	}
}

func renderDebuggerFooter(state diag.State) {
	panel.FillRectangle(0, 232, Width, 24, colorBackground)
	if state.LastError != "" {
		writeLargeUTF8(10, 250, "错误 "+state.LastError, colorRed)
	} else if state.CachePresent {
		writeLargeUTF8(10, 250, "缓存 OK", colorGreen)
	} else {
		writeLargeUTF8(10, 250, "无缓存", colorMuted)
	}
}
