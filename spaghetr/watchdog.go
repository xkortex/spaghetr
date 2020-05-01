package spaghetr

import (
	"context"
	"fmt"
	"sync"
	"time"
	"github.com/xkortex/vprint"
)

// A watchdog is a timer which can be reset
// implementations are *WatchdogCtx
type Watchdog interface {
	context.Context
	fmt.Stringer
	Stop()
	Feed()
}

// A whole bunch of copypasta since golib doesn't export these

// A canceler is a context type that can be canceled directly. The
// implementations are *cancelCtx and *timerCtx.
// from context/context.go
type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
}

// closedchan is a reusable closed channel.
// from context/context.go
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

// A cancelCtx can be canceled. When canceled, it also cancels any children
// that implement canceler.
// from context/context.go
type cancelCtx struct {
	context.Context

	mu       sync.Mutex            // protects following fields
	done     chan struct{}         // created lazily, closed by first cancel call
	children map[canceler]struct{} // set to nil by the first cancel call
	err      error                 // set to non-nil by the first cancel call
}

func (c *cancelCtx) Done() <-chan struct{} {
	c.mu.Lock()
	if c.done == nil {
		c.done = make(chan struct{})
	}
	d := c.done
	c.mu.Unlock()
	return d
}

func (c *cancelCtx) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}


// cancel closes c.done, cancels each of c's children, and, if
// removeFromParent is true, removes c from its parent's children.
// from context/context.go
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	vprint.Printf("cancelCtx cancel(%t): %v | err: %v\n", removeFromParent, c, err)
	if err == nil {
		panic("watchdog.go: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err
	if c.done == nil {
		vprint.Printf("c.done == nil\n")
		c.done = closedchan
	} else {
		vprint.Printf("close(c.done)\n")
		close(c.done)
	}
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		vprint.Printf("  cancelling: %v\n", child)
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		removeChild(c.Context, c)
	}
}

func (c *timerCtx) cancel(removeFromParent bool, err error) {
	vprint.Printf("timerCtx cancel(): %v\n")
	c.cancelCtx.cancel(false, err)
	if removeFromParent {
		// Remove this timerCtx from its parent cancelCtx's children.
		removeChild(c.cancelCtx.Context, c)
	}
	c.mu.Lock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.mu.Unlock()
}

// propagateCancel arranges for child to be canceled when parent is.
// from context/context.go
func propagateCancel(parent context.Context, child canceler) {
	vprint.Printf("propagateCancel(%v, %v)\n", parent, child)
	if parent.Done() == nil {
		return // parent is never canceled
	}
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		if p.err != nil {
			// parent has already been canceled
			child.cancel(false, p.err)
		} else {
			if p.children == nil {
				p.children = make(map[canceler]struct{})
			}
			p.children[child] = struct{}{}
		}
		p.mu.Unlock()
	} else {
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}

// parentCancelCtx follows a chain of parent references until it finds a
// *cancelCtx. This function understands how each of the concrete types in this
// package represents its parent.
// from context/context.go
func parentCancelCtx(parent context.Context) (*cancelCtx, bool) {
	vprint.Printf("parentCancelCtx(): %v\n", parent)

	for {
		switch c := parent.(type) {
		case *cancelCtx:
			vprint.Printf("cancel: cancelctx\n")
			return c, true
		case *timerCtx:
			vprint.Printf("cancel: timerctx\n")
			return &c.cancelCtx, true
		case *WatchdogCtx:
			vprint.Printf("cancel: watchdogctx\n")
			return &c.cancelCtx, true
		case *valueCtx:
			parent = c.Context
		default:
			vprint.Printf("cancel: default\n")
			return nil, false
		}
	}
}

// removeChild removes a context from its parent.
// from context/context.go
func removeChild(parent context.Context, child canceler) {
	p, ok := parentCancelCtx(parent)
	if !ok {
		return
	}
	p.mu.Lock()
	if p.children != nil {
		delete(p.children, child)
	}
	p.mu.Unlock()
}

// A timerCtx carries a timer and a deadline. It embeds a cancelCtx to
// implement Done and Err. It implements cancel by stopping its timer then
// delegating to cancelCtx.cancel.
// from context/context.go
type timerCtx struct {
	cancelCtx
	timer *time.Timer // Under cancelCtx.mu.

	deadline time.Time
}

// A valueCtx carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Context.
// only here to stay consistent with context.go
// from context/context.go
type valueCtx struct {
	context.Context
	key, val interface{}
}

// newCancelCtx returns an initialized cancelCtx.
// from context/context.go
func newCancelCtx(parent context.Context) cancelCtx {
	return cancelCtx{Context: parent}
}

type WatchdogCtx struct {
	cancelCtx
	interval        time.Duration
	timer           *time.Timer
	deadline        time.Time
	cancel_callback func()
}

func newWatchdog(parent context.Context, interval time.Duration) *WatchdogCtx {
	w := WatchdogCtx{
		cancelCtx: newCancelCtx(parent),
		interval:  interval,
		//timer: time.AfterFunc(interval, callback),
		deadline: time.Now().Add(interval),
	}
	w.cancel_callback = func() {
		vprint.Printf("watchdog expired\n")
		w.cancel(true, context.DeadlineExceeded)
	}
	return &w
}

func (c *WatchdogCtx) cancel(removeFromParent bool, err error) {
	vprint.Printf("WatchdogCtx cancel(%t): %v | err: %v\n", removeFromParent, c, err)
	c.cancelCtx.cancel(false, err)
	if removeFromParent {
		// Remove this timerCtx from its parent cancelCtx's children.
		removeChild(c.cancelCtx.Context, c)
	}
	c.mu.Lock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.mu.Unlock()
}

func (*WatchdogCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (w *WatchdogCtx) String() string {
	return "WatchdogCtx.WithInterval(" + w.interval.String() +
		 " [" + time.Until(w.deadline).String() + "])"
}

// Stop halts the watchdog timer without calling cancel()
// It will reactivate if Feed() is called
func (w *WatchdogCtx) Stop() {
	w.timer.Stop()
}

// Feed resets the timer associated with the watchdog, back to its set interval
// Calling Feed() on a stopped timer will restart it
func (w *WatchdogCtx) Feed() {
	vprint.Printf("Feed\n")
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timer == nil {
		vprint.Printf("Fed a dead dog x_x \n")
		return
	}
	w.timer.Stop()
	w.timer.Reset(w.interval)
	w.deadline = time.Now().Add(w.interval)
}

// WithWatchdog returns a copy of the parent with a watchdog timer attached.
// The returned context's Done channel is closed when the watchdog timer expires,
// when the returned cancel function is callled, or
// when the returned Done channel is closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithWatchdog(parent context.Context, interval time.Duration) (Watchdog, context.CancelFunc) {
	c := newWatchdog(parent, interval)

	propagateCancel(parent, c)

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err == nil {
		c.timer = time.AfterFunc(c.interval, c.cancel_callback)
	}
	return c, func() { c.cancel(true, context.Canceled) }
}
