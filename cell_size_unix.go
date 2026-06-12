//go:build !windows

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// getCellSize queries the terminal to get the exact width and height of a cell in pixels.
func getCellSize() (int, int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err == nil && ws.Xpixel > 0 && ws.Ypixel > 0 && ws.Col > 0 && ws.Row > 0 {
		cellW := int(ws.Xpixel / ws.Col)
		cellH := int(ws.Ypixel / ws.Row)
		if cellW > 0 && cellH > 0 {
			return cellW, cellH
		}
	}
	return 8, 16 // standard fallback estimate
}
