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
// // setWindowSize removes the title bar, sets the content size, and centers the
// // window within the screen's visible area. It must be used instead of
// // webview's SetSize because the webview library restores NSWindowStyleMaskTitled
// // before calling setContentSize:, which adds the title-bar height to the frame
// // and leaves the content area smaller than intended after setupWindow removes it.
// void setWindowSize(void* ptr, int width, int height) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     w.styleMask = NSWindowStyleMaskBorderless | NSWindowStyleMaskResizable;
//     w.opaque = NO;
//     w.backgroundColor = [NSColor clearColor];
//     w.level = NSFloatingWindowLevel;
//     if ([w.contentView isKindOfClass:[WKWebView class]]) {
//         WKWebView* webView = (WKWebView*)w.contentView;
//         [webView setValue:@NO forKey:@"drawsBackground"];
//     }
//     [w setContentSize:NSMakeSize(width, height)];
//     NSScreen* screen = w.screen ? w.screen : [NSScreen mainScreen];
//     NSRect visible = screen.visibleFrame;
//     NSSize ws = w.frame.size;
//     NSPoint origin = NSMakePoint(
//         NSMidX(visible) - ws.width / 2.0,
//         NSMidY(visible) - ws.height / 2.0
//     );
//     if (origin.y < NSMinY(visible)) origin.y = NSMinY(visible);
//     if (origin.x < NSMinX(visible)) origin.x = NSMinX(visible);
//     [w setFrameOrigin:origin];
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

func SetWindowSize(window unsafe.Pointer, width, height int) {
	C.setWindowSize(window, C.int(width), C.int(height))
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
