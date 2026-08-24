package clock

import (
	"time"

	"github.com/lacsar712/hydrturn/internal/model"
)

type Window struct{ Duration time.Duration }

func NewWindow(d time.Duration) Window {
	if d < 0 {
		d = 0
	}
	return Window{Duration: d}
}

func (w Window) Satisfied(clk ProcessClock, anchor time.Time) bool {
	if w.Duration == 0 {
		return true
	}
	return clk.Since(anchor) >= w.Duration
}

func (w Window) WaitUntil(clk ProcessClock, anchor time.Time) error {
	if w.Satisfied(clk, anchor) {
		return nil
	}
	return model.ErrWindowOpen
}

type SyncWindow struct {
	clk    ProcessClock
	window Window
}

func NewSyncWindow(clk ProcessClock) *SyncWindow {
	return &SyncWindow{clk: clk, window: NewWindow(model.SyncWindow)}
}

func (p *SyncWindow) Ready(startedAt time.Time) bool {
	return time.Since(startedAt) >= model.SyncWindow
}

func (p *SyncWindow) Require(startedAt time.Time) error {
	if p.Ready(startedAt) {
		return nil
	}
	return model.ErrSyncIncomplete
}

type IgnitionDelayWindow struct {
	clk    ProcessClock
	window Window
}

func NewIgnitionDelayWindow(clk ProcessClock) *IgnitionDelayWindow {
	return &IgnitionDelayWindow{clk: clk, window: NewWindow(model.IgnitionDelayWindow)}
}

func (i *IgnitionDelayWindow) Ready(ignitionAt time.Time) bool {
	return i.window.Satisfied(i.clk, ignitionAt)
}

func (i *IgnitionDelayWindow) Require(ignitionAt time.Time) error {
	if i.Ready(ignitionAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type DraftubeSwellWindow struct {
	clk    ProcessClock
	window Window
}

func NewDraftubeSwellWindow(clk ProcessClock) *DraftubeSwellWindow {
	return &DraftubeSwellWindow{clk: clk, window: NewWindow(model.DraftubeSwellSettleWindow)}
}

func (d *DraftubeSwellWindow) Settled(swellAt time.Time) bool {
	return d.window.Satisfied(d.clk, swellAt)
}

func (d *DraftubeSwellWindow) RequireSettled(swellAt time.Time) error {
	if d.Settled(swellAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type WicketWarmupWindow struct {
	clk    ProcessClock
	window Window
}

func NewWicketWarmupWindow(clk ProcessClock) *WicketWarmupWindow {
	return &WicketWarmupWindow{clk: clk, window: NewWindow(model.WicketWarmupWindow)}
}

func (c *WicketWarmupWindow) Ready(ignitionAt time.Time) bool {
	return c.window.Satisfied(c.clk, ignitionAt)
}

func (c *WicketWarmupWindow) Require(ignitionAt time.Time) error {
	if c.Ready(ignitionAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type FeedwaterRampTracker struct {
	clk    ProcessClock
	window Window
	anchor time.Time
}

func NewFeedwaterRampTracker(clk ProcessClock) *FeedwaterRampTracker {
	return &FeedwaterRampTracker{
		clk:    clk,
		window: NewWindow(model.FeedwaterRampWindow),
		anchor: clk.Now(),
	}
}

func (f *FeedwaterRampTracker) Reset()          { f.anchor = f.clk.Now() }
func (f *FeedwaterRampTracker) Satisfied() bool { return f.window.Satisfied(f.clk, f.anchor) }
func (f *FeedwaterRampTracker) Require() error  { return f.window.WaitUntil(f.clk, f.anchor) }
