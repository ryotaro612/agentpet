package internal

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa -framework WebKit
// #import <Cocoa/Cocoa.h>
// #import <WebKit/WebKit.h>
//
// void hideToolbar(void* ptr) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     w.styleMask |= NSWindowStyleMaskFullSizeContentView;
//     w.titlebarAppearsTransparent = YES;
//     w.titleVisibility = NSWindowTitleHidden;
//     [w standardWindowButton:NSWindowCloseButton].hidden = YES;
//     [w standardWindowButton:NSWindowMiniaturizeButton].hidden = YES;
//     [w standardWindowButton:NSWindowZoomButton].hidden = YES;
// }
//
// static WKWebView* findWebView(NSView* view) {
//     if ([view isKindOfClass:[WKWebView class]]) return (WKWebView*)view;
//     for (NSView* sub in view.subviews) {
//         WKWebView* found = findWebView(sub);
//         if (found) return found;
//     }
//     return nil;
// }
//
// void makeTransparent(void* ptr) {
//     NSWindow* w = (__bridge NSWindow*)ptr;
//     w.opaque = NO;
//     w.backgroundColor = NSColor.clearColor;
//     WKWebView* webView = findWebView(w.contentView);
//     if (webView) {
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
import "C"
import "unsafe"

func HideToolbar(window unsafe.Pointer) {
	C.hideToolbar(window)
}

func MakeTransparent(window unsafe.Pointer) {
	C.makeTransparent(window)
}

func MoveWindow(window unsafe.Pointer, dx, dy float64) {
	C.moveWindow(window, C.double(dx), C.double(dy))
}
