//go:build windows

package main

// getCellSize returns the fallback terminal cell size for Windows.
func getCellSize() (int, int) {
	return 8, 16 // standard fallback estimate
}
