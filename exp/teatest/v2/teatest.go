// Package teatest provides helper functions to test tea.Model's.
package teatest

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/vt"
)

// Program defines the subset of the tea.Program API we need for testing.
type Program interface {
	Send(tea.Msg)
}

// TestModelOptions defines all options available to the test function.
type TestModelOptions struct {
	size tea.WindowSizeMsg
}

// TestOption is a functional option.
type TestOption func(opts *TestModelOptions)

// WithInitialTermSize ...
func WithInitialTermSize(x, y int) TestOption {
	return func(opts *TestModelOptions) {
		opts.size = tea.WindowSizeMsg{
			Width:  x,
			Height: y,
		}
	}
}

// WaitingForContext is the context for a WaitFor.
type WaitingForContext struct {
	Duration      time.Duration
	CheckInterval time.Duration
}

// WaitForOption changes how a WaitFor will behave.
type WaitForOption func(*WaitingForContext)

// WithCheckInterval sets how much time a WaitFor should sleep between every
// check.
func WithCheckInterval(d time.Duration) WaitForOption {
	return func(wf *WaitingForContext) {
		wf.CheckInterval = d
	}
}

// WithDuration sets how much time a WaitFor will wait for the condition.
func WithDuration(d time.Duration) WaitForOption {
	return func(wf *WaitingForContext) {
		wf.Duration = d
	}
}

// WaitForOutput keeps reading from r until the condition matches.
// Default duration is 1s, default check interval is 50ms.
// These defaults can be changed with WithDuration and WithCheckInterval.
func WaitForOutput(
	tb testing.TB,
	tm *TestModel,
	condition func(string) bool,
	options ...WaitForOption,
) {
	tb.Helper()
	if err := doWaitFor(tm, condition, options...); err != nil {
		tb.Fatal(err)
	}
}

func doWaitFor(tm *TestModel, condition func(string) bool, options ...WaitForOption) error {
	wf := WaitingForContext{
		Duration:      time.Second,
		CheckInterval: 50 * time.Millisecond, //nolint: gomnd
	}

	for _, opt := range options {
		opt(&wf)
	}

	start := time.Now()
	for time.Since(start) <= wf.Duration {
		if condition(tm.Output()) {
			return nil
		}
		time.Sleep(wf.CheckInterval)
	}
	return fmt.Errorf("WaitFor: condition not met after %s. Last output:\n%q", wf.Duration, tm.Output())
}

// TestModel is a model that is being tested.
type TestModel struct {
	program *tea.Program

	term *vt.Terminal

	modelCh chan tea.Model
	model   tea.Model

	done   sync.Once
	doneCh chan bool
}

// NewTestModel makes a new TestModel which can be used for tests.
func NewTestModel(tb testing.TB, m tea.Model, options ...TestOption) *TestModel {
	var opts TestModelOptions
	for _, opt := range options {
		opt(&opts)
	}
	if opts.size.Width == 0 {
		opts.size.Width, opts.size.Height = 70, 40
	}

	tm := &TestModel{
		term:    vt.NewTerminal(opts.size.Width, opts.size.Height),
		modelCh: make(chan tea.Model, 1),
		doneCh:  make(chan bool, 1),
	}

	tm.program = tea.NewProgram(
		m,
		tea.WithInput(tm.term),
		tea.WithOutput(tm.term),
		tea.WithoutSignals(),
	)

	interruptions := make(chan os.Signal, 1)
	signal.Notify(interruptions, syscall.SIGINT)
	go func() {
		m, err := tm.program.Run()
		if err != nil {
			tb.Fatalf("app failed: %s", err)
		}
		tm.doneCh <- true
		tm.modelCh <- m
	}()
	go func() {
		<-interruptions
		signal.Stop(interruptions)
		tb.Log("interrupted")
		tm.program.Kill()
	}()

	tm.program.Send(opts.size)
	return tm
}

func mergeOpts(opts []FinalOpt) FinalOpts {
	r := FinalOpts{}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func (tm *TestModel) waitDone(tb testing.TB, opts FinalOpts) {
	tm.done.Do(func() {
		if opts.timeout > 0 {
			select {
			case <-time.After(opts.timeout):
				if opts.onTimeout == nil {
					tb.Fatalf("timeout after %s", opts.timeout)
				}
				opts.onTimeout(tb)
			case <-tm.doneCh:
			}
		} else {
			<-tm.doneCh
		}
	})
}

// FinalOpts represents the options for FinalModel and FinalOutput.
type FinalOpts struct {
	timeout   time.Duration
	onTimeout func(tb testing.TB)
	trim      bool
}

// FinalOpt changes FinalOpts.
type FinalOpt func(opts *FinalOpts)

// WithTimeoutFn allows to define what happens when WaitFinished times out.
func WithTimeoutFn(fn func(tb testing.TB)) FinalOpt {
	return func(opts *FinalOpts) {
		opts.onTimeout = fn
	}
}

// WithFinalTimeout allows to set a timeout for how long FinalModel and
// FinalOuput should wait for the program to complete.
func WithFinalTimeout(d time.Duration) FinalOpt {
	return func(opts *FinalOpts) {
		opts.timeout = d
	}
}

// WaitFinished waits for the app to finish.
// This method only returns once the program has finished running or when it
// times out.
func (tm *TestModel) WaitFinished(tb testing.TB, opts ...FinalOpt) {
	tm.waitDone(tb, mergeOpts(opts))
}

// FinalModel returns the resulting model, resulting from program.Run().
// This method only returns once the program has finished running or when it
// times out.
func (tm *TestModel) FinalModel(tb testing.TB, opts ...FinalOpt) tea.Model {
	tm.WaitFinished(tb, opts...)
	select {
	case m := <-tm.modelCh:
		tm.model = m
		return tm.model
	default:
		return tm.model
	}
}

// FinalOutput returns the program's final output.
// This method only returns once the program has finished running or when it
// times out.
// It's the equivalent of calling both `tm.WaitFinished` and `tm.Output()`.
func (tm *TestModel) FinalOutput(tb testing.TB, opts ...FinalOpt) string {
	tm.WaitFinished(tb, opts...)
	return tm.Output()
}

// Output returns the program's current output.
func (tm *TestModel) Output() string {
	return tm.term.String()
}

// Send sends messages to the underlying program.
func (tm *TestModel) Send(m tea.Msg) {
	tm.program.Send(m)
}

// Quit quits the program and releases the terminal.
func (tm *TestModel) Quit() error {
	tm.program.Quit()
	return nil
}

// Type types the given text into the given program.
func (tm *TestModel) Type(s string) {
	for _, c := range s {
		tm.Send(tea.KeyPressMsg{
			Code: c,
			Text: string(c),
		})
	}
}

// GetProgram gets the TestModel's program.
func (tm *TestModel) GetProgram() *tea.Program {
	return tm.program
}

// RequireEqualOutput is a helper function to assert the given output is
// the expected from the golden files, printing its diff in case it is not.
//
// Important: this uses the system `diff` tool.
//
// You can update the golden files by running your tests with the -update flag.
func RequireEqualOutput(tb testing.TB, out string) {
	tb.Helper()
	golden.RequireEqualEscape(tb, []byte(out), true)
}

// TrimEmptyLines removes trailing empty lines from the given output.
func TrimEmptyLines(out string) string {
	// trim empty trailing lines from the output
	lines := strings.Split(out, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.Join(lines[:i], "\n")
		}
	}
	return out
}
