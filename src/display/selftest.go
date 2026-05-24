package display

import "time"

func RenderSelfTest() {
	if !ready {
		return
	}

	panel.FillScreen(colorRed)
	time.Sleep(250 * time.Millisecond)
	panel.FillScreen(colorGreen)
	time.Sleep(250 * time.Millisecond)
	panel.FillScreen(colorBlue)
	time.Sleep(250 * time.Millisecond)

	panel.FillScreen(colorBackground)
	panel.FillRectangle(0, 0, 60, 240, colorRed)
	panel.FillRectangle(60, 0, 60, 240, colorGreen)
	panel.FillRectangle(120, 0, 60, 240, colorBlue)
	panel.FillRectangle(180, 0, 60, 240, colorText)
	panel.FillRectangle(0, 0, 240, 4, colorYellow)
	panel.FillRectangle(0, 236, 240, 4, colorYellow)
	panel.FillRectangle(0, 0, 4, 240, colorYellow)
	panel.FillRectangle(236, 0, 4, 240, colorYellow)
	writeSmall(14, 34, "ST7789 TEST", colorText)
	writeTiny(14, 52, "SCK14 SDA13 DC7 RST6", colorText)
	time.Sleep(2500 * time.Millisecond)
}
