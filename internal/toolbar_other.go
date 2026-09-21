//go:build !darwin

package internal

import "unsafe"

func HideToolbar(_ unsafe.Pointer)                {}
func MakeTransparent(_ unsafe.Pointer)            {}
func MoveWindow(_ unsafe.Pointer, _, _ float64) {}
