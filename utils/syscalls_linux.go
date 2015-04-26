// +build linux

package utils

import (
	"time"
)

func sleep(timeout time.Duration) error {
	return nil
}

func openFileMapping(lpName string) (uintptr, error) {
	return 0, nil
}

func mapViewOfFile(hMemMapFile uintptr, dwNumberOfBytesToMap int) (uintptr, error) {
	return 0, nil
}

func closeHandle(handle uintptr) error {
	return nil
}

func unmapViewOfFile(lpBaseAddress uintptr) error {
	return nil
}

func openEvent(lpName string) (uintptr, error) {
	return 0, nil
}

func waitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
	return nil
}

func registerWindowMessageA(lpString string) (uint, error) {
	return 0, nil
}

func sendNotifyMessage(msgID uint, wParam uint32, lParam uint32) error {
	return nil
}

func MAKELONG(lo, hi uint16) uint32 {
	return 0
}

func now() time.Time {
	return time.Now()
}
