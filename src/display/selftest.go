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
	stripeWidth := Width / 4
	panel.FillRectangle(0, 0, stripeWidth, Height, colorRed)
	panel.FillRectangle(stripeWidth, 0, stripeWidth, Height, colorGreen)
	panel.FillRectangle(stripeWidth*2, 0, stripeWidth, Height, colorBlue)
	panel.FillRectangle(stripeWidth*3, 0, Width-stripeWidth*3, Height, colorText)
	panel.FillRectangle(0, 0, Width, 4, colorYellow)
	panel.FillRectangle(0, Height-4, Width, 4, colorYellow)
	panel.FillRectangle(0, 0, 4, Height, colorYellow)
	panel.FillRectangle(Width-4, 0, 4, Height, colorYellow)
	writeSmall(14, 34, "ST7789V2 240x320", colorText)
	writeTiny(14, 52, "CS10 DC11 RST12", colorText)
	writeTiny(14, 64, "SCL13 SDA14 BL=3V3", colorText)
	time.Sleep(2500 * time.Millisecond)
}
