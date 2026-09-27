package view

import "github.com/webview/webview"

type internalView interface {
	SetUpWIndow()
	PreventTerminateOnHide()
	PreventHide()
	SetupQuit()
	MoveWindow(dx, dy float64)
	SHowWindow()
	HideWindow()
	SetWindowSize(width, height int)
	ResizeWindowKeepingPosition(width, height int)
	Bind(name string, f any) error
	Navigate(url string)
	Dispatch(f func())
	Terminate()
	Run()
	Destroy()
	Eval(js string)
}

type webviewAdapter struct {
	wv webview.WebView
}

func NewWebviewAdapter(wv webview.WebView) *webviewAdapter {
	return &webviewAdapter{wv: wv}
}

func (a *webviewAdapter) SetUpWIndow()              { SetupWindow(a.wv.Window()) }
func (a *webviewAdapter) PreventTerminateOnHide()   { PreventTerminateOnHide() }
func (a *webviewAdapter) PreventHide()              { PreventHide(a.wv.Window()) }
func (a *webviewAdapter) SetupQuit()                                    { SetupQuit() }
func (a *webviewAdapter) MoveWindow(dx, dy float64)                     { MoveWindow(a.wv.Window(), dx, dy) }
func (a *webviewAdapter) SHowWindow()                                   { win := a.wv.Window(); a.wv.Dispatch(func() { ShowWindow(win) }) }
func (a *webviewAdapter) HideWindow()                                   { win := a.wv.Window(); a.wv.Dispatch(func() { HideWindow(win) }) }
func (a *webviewAdapter) SetWindowSize(width, height int)               { SetWindowSize(a.wv.Window(), width, height) }
func (a *webviewAdapter) ResizeWindowKeepingPosition(width, height int) { ResizeWindowKeepingPosition(a.wv.Window(), width, height) }
func (a *webviewAdapter) Bind(name string, f any) error                 { return a.wv.Bind(name, f) }
func (a *webviewAdapter) Navigate(url string)                           { a.wv.Navigate(url) }
func (a *webviewAdapter) Dispatch(f func())                             { a.wv.Dispatch(f) }
func (a *webviewAdapter) Terminate()                                    { a.wv.Terminate() }
func (a *webviewAdapter) Run()                                          { a.wv.Run() }
func (a *webviewAdapter) Destroy()                                      { a.wv.Destroy() }
func (a *webviewAdapter) Eval(js string)                                { a.wv.Eval(js) }
