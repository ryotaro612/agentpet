//go:build !darwin

package internal

import "unsafe"

func SetupWindow(_ unsafe.Pointer)              {}
func MoveWindow(_ unsafe.Pointer, _, _ float64) {}
