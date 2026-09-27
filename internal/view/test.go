package view

// NoopInternalView implements internalView with all methods as no-ops.
// Embed it in test types to avoid implementing every method when only a few
// are needed.
type NoopInternalView struct{}

func (NoopInternalView) PreventTerminateOnHide()              {}
func (NoopInternalView) SetUpWIndow()                         {}
func (NoopInternalView) PreventHide()                         {}
func (NoopInternalView) SetupQuit()                           {}
func (NoopInternalView) MoveWindow(_, _ float64)              {}
func (NoopInternalView) SHowWindow()                          {}
func (NoopInternalView) HideWindow()                          {}
func (NoopInternalView) SetWindowSize(_, _ int)               {}
func (NoopInternalView) ResizeWindowKeepingPosition(_, _ int) {}
func (NoopInternalView) Bind(_ string, _ any) error           { return nil }
func (NoopInternalView) Navigate(_ string)                    {}
func (NoopInternalView) Terminate()                           {}
func (NoopInternalView) Run()                                 {}
func (NoopInternalView) Destroy()                             {}
func (NoopInternalView) Eval(_ string)                        {}
func (NoopInternalView) Dispatch(f func())                    { f() }

// MockInternalView records calls to SetWindowSize, ResizeWindowKeepingPosition,
// Eval, and Dispatch. All other methods delegate to NoopInternalView.
// DispatchDone receives a value after each Dispatch closure completes, letting
// tests synchronise after a known number of animations.
type MockInternalView struct {
	NoopInternalView
	SetWindowSizeCalls [][2]int
	ResizeWindowCalls  [][2]int
	EvalCalls          []string
	DispatchDone       chan struct{}
}

func NewMockInternalView() *MockInternalView {
	return &MockInternalView{DispatchDone: make(chan struct{}, 16)}
}

func (m *MockInternalView) SetWindowSize(w, h int) {
	m.SetWindowSizeCalls = append(m.SetWindowSizeCalls, [2]int{w, h})
}

func (m *MockInternalView) ResizeWindowKeepingPosition(w, h int) {
	m.ResizeWindowCalls = append(m.ResizeWindowCalls, [2]int{w, h})
}

func (m *MockInternalView) Eval(js string) {
	m.EvalCalls = append(m.EvalCalls, js)
}

func (m *MockInternalView) Dispatch(f func()) {
	f()
	m.DispatchDone <- struct{}{}
}
