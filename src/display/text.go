package display

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/proggy"
)

func writeTitle(x, y int16, text string, c color.RGBA) {
	tinyfont.WriteLine(&panel, &freemono.Regular9pt7b, x, y, asciiOnly(text, 18), c)
}

func writeSmall(x, y int16, text string, c color.RGBA) {
	tinyfont.WriteLine(&panel, &proggy.TinySZ8pt7b, x, y, asciiOnly(text, 36), c)
}

func writeTiny(x, y int16, text string, c color.RGBA) {
	tinyfont.WriteLine(&panel, &tinyfont.TomThumb, x, y, asciiOnly(text, 46), c)
}

func asciiOnly(s string, maxRunes int) string {
	var out strings.Builder
	count := 0
	for len(s) > 0 && count < maxRunes {
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && size == 0 {
			break
		}
		if r >= 32 && r <= 126 {
			out.WriteByte(byte(r))
			count++
		} else if r == '\n' || r == '\r' || r == '\t' {
			out.WriteByte(' ')
			count++
		} else {
			out.WriteByte('?')
			count++
		}
		s = s[size:]
	}
	return out.String()
}

func trimLogPrefix(s string) string {
	const prefix = "[tinygo-setup] "
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}
