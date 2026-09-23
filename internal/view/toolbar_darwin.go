//go:build darwin

package view

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa -framework WebKit
// #import <Cocoa/Cocoa.h>
// #import <WebKit/WebKit.h>
//
// void setupWindow(void* ptr) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     w.styleMask = NSWindowStyleMaskBorderless | NSWindowStyleMaskResizable;
//     w.opaque = NO;
//     w.backgroundColor = [NSColor clearColor];
//     w.level = NSFloatingWindowLevel;
//     if ([w.contentView isKindOfClass:[WKWebView class]]) {
//         WKWebView* webView = (WKWebView*)w.contentView;
//         [webView setValue:@NO forKey:@"drawsBackground"];
//     }
// }
//
// void moveWindow(void* ptr, double dx, double dy) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     NSPoint o = w.frame.origin;
//     o.x += dx;
//     o.y -= dy;
//     [w setFrameOrigin:o];
// }
//
// void showWindowNative(void* ptr) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     [w makeKeyAndOrderFront:nil];
// }
//
// void hideWindowNative(void* ptr) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     [w orderOut:nil];
// }
import "C"
import "unsafe"

func SetupWindow(window unsafe.Pointer) {
	C.setupWindow(window)
}

func MoveWindow(window unsafe.Pointer, dx, dy float64) {
	C.moveWindow(window, C.double(dx), C.double(dy))
}

func ShowWindow(window unsafe.Pointer) {
	C.showWindowNative(window)
}

func HideWindow(window unsafe.Pointer) {
	C.hideWindowNative(window)
}
