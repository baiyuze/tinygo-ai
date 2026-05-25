package display

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/freesans"
	"tinygo.org/x/tinyfont/notoemoji"
	"tinygo.org/x/tinyfont/proggy"
)

func writeTitle(x, y int16, text string, c color.RGBA) {
	tinyfont.WriteLine(&panel, &freemono.Regular9pt7b, x, y, asciiOnly(text, 18), c)
}

func writeSmall(x, y int16, text string, c color.RGBA) {
	tinyfont.WriteLine(&panel, &proggy.TinySZ8pt7b, x, y, asciiOnly(text, 36), c)
}

func writeSmallUTF8(x, y int16, text string, c color.RGBA) {
	writeUTF8(x, y, text, c, false)
}

func writeLargeUTF8(x, y int16, text string, c color.RGBA) {
	writeUTF8(x, y, text, c, true)
}

func writeUTF8(x, y int16, text string, c color.RGBA, large bool) {
	const gap = int16(1)

	maxX := Width - 6
	cjkTop := y - cjkGlyphHeight + 4
	asciiRun := strings.Builder{}

	flushASCII := func() {
		if asciiRun.Len() == 0 {
			return
		}
		s := asciiRun.String()
		if large && cjkGlyphHeight > 12 {
			tinyfont.WriteLine(&panel, &freesans.Bold12pt7b, x, y, asciiOnly(s, 24), c)
			x += estimateLargeASCIIWidth(s)
		} else {
			tinyfont.WriteLine(&panel, &proggy.TinySZ8pt7b, x, y, asciiOnly(s, 36), c)
			x += int16(len(s)) * 6
		}
		asciiRun.Reset()
	}

	for _, r := range text {
		if x >= maxX {
			break
		}
		if r >= 32 && r <= 126 {
			asciiRun.WriteByte(byte(r))
			continue
		}
		if isVariationSelector(r) {
			continue
		}

		flushASCII()
		if x+cjkGlyphWidth > maxX {
			break
		}
		if glyphIndex, ok := lookupCJKGlyph(r); ok {
			drawCJKGlyph(x, cjkTop, glyphIndex, c)
			x += cjkGlyphWidth + gap
		} else if isEmojiRune(r) {
			tinyfont.WriteLine(&panel, &notoemoji.NotoEmojiRegular20pt, x, y+2, string(r), c)
			x += 26
		} else {
			drawMissingGlyph(x, cjkTop, c)
			x += cjkGlyphWidth + gap
		}
	}
	flushASCII()
}

func estimateLargeASCIIWidth(s string) int16 {
	var width int16
	for _, r := range s {
		switch {
		case r == ' ':
			width += 6
		case r >= '0' && r <= '9':
			width += 11
		case r >= 'A' && r <= 'Z':
			width += 13
		case r >= 'a' && r <= 'z':
			width += 11
		default:
			width += 8
		}
	}
	return width
}

func isVariationSelector(r rune) bool {
	return r >= 0xFE00 && r <= 0xFE0F
}

func isEmojiRune(r rune) bool {
	return (r >= 0x2600 && r <= 0x27BF) ||
		(r >= 0x1F300 && r <= 0x1FAFF)
}

func writeTiny(x, y int16, text string, c color.RGBA) {
	tinyfont.WriteLine(&panel, &tinyfont.TomThumb, x, y, asciiOnly(text, 46), c)
}

func drawCJKGlyph(x, y int16, glyphIndex int, c color.RGBA) {
	for row := int16(0); row < cjkGlyphHeight; row++ {
		bits := cjkGlyphRow(glyphIndex, row)
		runStart := int16(-1)
		for col := int16(0); col < cjkGlyphWidth; col++ {
			mask := uint32(1) << uint(cjkGlyphWidth-1-col)
			if bits&mask != 0 {
				if runStart < 0 {
					runStart = col
				}
				continue
			}
			if runStart >= 0 {
				panel.FillRectangle(x+runStart, y+row, col-runStart, 1, c)
				runStart = -1
			}
		}
		if runStart >= 0 {
			panel.FillRectangle(x+runStart, y+row, cjkGlyphWidth-runStart, 1, c)
		}
	}
}

func drawMissingGlyph(x, y int16, c color.RGBA) {
	panel.FillRectangle(x, y, cjkGlyphWidth, 1, c)
	panel.FillRectangle(x, y+cjkGlyphHeight-1, cjkGlyphWidth, 1, c)
	panel.FillRectangle(x, y, 1, cjkGlyphHeight, c)
	panel.FillRectangle(x+cjkGlyphWidth-1, y, 1, cjkGlyphHeight, c)
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
