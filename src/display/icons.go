package display

import "image/color"

func drawWiFiIcon(x, y int16, c color.RGBA) {
	drawSignalBar(x, y+13, 3, c)
	drawSignalBar(x+5, y+9, 7, c)
	drawSignalBar(x+10, y+5, 11, c)
	drawSignalBar(x+15, y+1, 15, c)
}

func drawWiFiIconLarge(x, y int16, c color.RGBA) {
	drawSignalBarLarge(x, y+18, 5, c)
	drawSignalBarLarge(x+8, y+13, 10, c)
	drawSignalBarLarge(x+16, y+8, 15, c)
	drawSignalBarLarge(x+24, y+3, 20, c)
}

func drawSignalBar(x, y, h int16, c color.RGBA) {
	panel.FillRectangle(x, y, 3, h, c)
}

func drawSignalBarLarge(x, y, h int16, c color.RGBA) {
	panel.FillRectangle(x, y, 5, h, c)
}

func drawStatusDot(x, y int16, c color.RGBA) {
	panel.FillRectangle(x, y, 6, 6, c)
	panel.FillRectangle(x+1, y-1, 4, 8, c)
}

func drawStatusDotLarge(x, y int16, c color.RGBA) {
	panel.FillRectangle(x, y, 12, 12, c)
	panel.FillRectangle(x+2, y-2, 8, 16, c)
}

func drawPanel(x, y, w, h int16) {
	panel.FillRectangle(x, y, w, h, colorPanel)
	panel.FillRectangle(x, y, w, 1, colorBorder)
	panel.FillRectangle(x, y+h-1, w, 1, colorBorder)
	panel.FillRectangle(x, y, 1, h, colorBorder)
	panel.FillRectangle(x+w-1, y, 1, h, colorBorder)
}
