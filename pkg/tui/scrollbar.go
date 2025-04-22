package tui

import (
	"math"
	"strings"
)

type ScrollbarStyleInterface interface {
	GetScrollbarThumb() string
	GetScrollbarTrack() string
	GetScrollbarWidth() int
}

func Scrollbar(
	style ScrollbarStyleInterface,
	height, total, visible, offset int,
) string {
	if total == visible {
		return strings.TrimRight(strings.Repeat(" \n", height), "\n")
	}
	ratio := float64(height) / float64(total)
	thumbHeight := max(1, int(math.Round(float64(visible)*ratio)))
	thumbOffset := max(0, min(height-thumbHeight, int(math.Round(float64(offset)*ratio))))

	return strings.TrimRight(
		strings.Repeat(style.GetScrollbarTrack()+"\n", thumbOffset)+
			strings.Repeat(style.GetScrollbarThumb()+"\n", thumbHeight)+
			strings.Repeat(style.GetScrollbarTrack()+"\n", max(0, height-thumbOffset-thumbHeight)),
		"\n",
	)
}
