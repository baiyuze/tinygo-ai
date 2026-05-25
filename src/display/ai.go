package display

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	aiMaxLines      = 14
	aiMaxLineRunes  = 17
	aiLineHeight    = int16(17)
	aiTextStartY    = int16(56)
	aiScreenTimeout = 5 * time.Minute
)

var aiScreenUntil time.Time

func RenderAIText(text string) {
	if !ready {
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	aiScreenUntil = time.Now().Add(aiScreenTimeout)
	panel.FillScreen(colorBackground)
	writeLargeUTF8(10, 24, "AI", colorBlue)
	writeSmall(40, 22, "response", colorMuted)
	panel.FillRectangle(10, 36, Width-20, 1, colorBorder)

	lines := wrapDisplayText(text, aiMaxLines, aiMaxLineRunes)
	y := aiTextStartY
	for _, line := range lines {
		writeLargeUTF8(10, y, line, colorText)
		y += aiLineHeight
	}
}

func aiScreenActive() bool {
	return !aiScreenUntil.IsZero() && time.Now().Before(aiScreenUntil)
}

func wrapDisplayText(text string, maxLines, maxRunes int) []string {
	lines := make([]string, 0, maxLines)
	var line strings.Builder
	count := 0

	flush := func() {
		if line.Len() == 0 || len(lines) >= maxLines {
			line.Reset()
			count = 0
			return
		}
		lines = append(lines, strings.TrimSpace(line.String()))
		line.Reset()
		count = 0
	}

	for len(text) > 0 && len(lines) < maxLines {
		r, size := utf8.DecodeRuneInString(text)
		if r == utf8.RuneError && size == 0 {
			break
		}
		text = text[size:]

		if r == '\r' {
			continue
		}
		if r == '\n' {
			flush()
			continue
		}
		if count >= maxRunes {
			flush()
		}
		line.WriteRune(r)
		count++
	}
	if len(lines) < maxLines {
		flush()
	}
	return lines
}
