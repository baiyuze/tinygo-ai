package display

func RenderMicVerify(level int, leftPeak, rightPeak uint32, minSample, maxSample int32, toggles int) {
	if !ready {
		return
	}
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}

	panel.FillScreen(colorBackground)
	writeTitle(12, 28, "INMP441 MIC TEST", colorText)
	writeSmall(12, 50, "SCK GPIO6  WS GPIO5  SD GPIO4", colorMuted)
	writeSmall(12, 64, "LR=GND LEFT CHANNEL", colorMuted)

	writeLargeUTF8(12, 98, "LEVEL", colorBlue)
	drawMicLevelBar(12, 112, 216, 28, level)
	writeSmall(12, 156, "LEFT PEAK  "+uintText(leftPeak), colorText)
	writeSmall(12, 170, "RIGHT PEAK "+uintText(rightPeak), colorMuted)
	writeSmall(12, 184, "MIN  "+intText(minSample), colorMuted)
	writeSmall(12, 198, "MAX  "+intText(maxSample), colorMuted)
	writeSmall(12, 212, "SD TOGGLES "+intText(int32(toggles)), colorMuted)
	writeSmall(12, 226, "EXPECT LEFT ACTIVE", colorMuted)

	if rightPeak > 8000 && leftPeak < 1000 {
		writeLargeUTF8(12, 252, "CHECK LR/SD", colorRed)
	} else if leftPeak > 8000 && toggles > 100 {
		writeLargeUTF8(12, 252, "LEFT DATA ACTIVE", colorGreen)
	} else if toggles > 100 {
		writeLargeUTF8(12, 252, "CLOCK/DATA ONLY", colorYellow)
	} else {
		writeLargeUTF8(12, 252, "NO I2S DATA", colorRed)
	}
}

func drawMicLevelBar(x, y, w, h int16, level int) {
	panel.FillRectangle(x, y, w, h, colorPanel)
	panel.FillRectangle(x, y, w, 1, colorBorder)
	panel.FillRectangle(x, y+h-1, w, 1, colorBorder)
	panel.FillRectangle(x, y, 1, h, colorBorder)
	panel.FillRectangle(x+w-1, y, 1, h, colorBorder)

	fill := int16(int(w-4) * level / 100)
	c := colorGreen
	if level < 10 {
		c = colorYellow
	}
	if level > 80 {
		c = colorRed
	}
	if fill > 0 {
		panel.FillRectangle(x+2, y+2, fill, h-4, c)
	}
}

func uintText(value uint32) string {
	if value == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

func intText(value int32) string {
	if value < 0 {
		return "-" + uintText(uint32(-value))
	}
	return uintText(uint32(value))
}
