//go:build !darwin

package view

import "unsafe"

func SetupWindow(_ unsafe.Pointer)                           {}
func PreventHide(_ unsafe.Pointer)                           {}
func SetWindowSize(_ unsafe.Pointer, _, _ int)               {}
func ResizeWindowKeepingPosition(_ unsafe.Pointer, _, _ int) {}
func MoveWindow(_ unsafe.Pointer, _, _ float64)              {}
func SetupQuit()                                             {}
func ShowWindow(_ unsafe.Pointer)                            {}
func HideWindow(_ unsafe.Pointer)                            {}
func PreventTerminateOnHide()                                {}
