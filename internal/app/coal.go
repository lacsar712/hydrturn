package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/hydrturn/internal/clock"
	"github.com/lacsar712/hydrturn/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

var activeBladeCancel context.CancelFunc

func (a *App) bindBladeLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if activeBladeCancel != nil {
		activeBladeCancel()
	}
	child, cancel := context.WithCancel(ctx)
	activeBladeCancel = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelBladeLoop(holder string) {
	a.mu.Lock()
	if activeBladeCancel != nil {
		activeBladeCancel()
		activeBladeCancel = nil
	}
	a.mu.Unlock()
}

func (a *App) cancelAllBladeLoops() {
	a.mu.Lock()
	for holder, cancel := range a.bladeLoopCancels {
		cancel()
		delete(a.bladeLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Wicket.BladeFlowTPH
}

func (a *App) RunBladeRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindBladeLoop(holder, ctx)
	defer a.cancelBladeLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Wicket.BladeFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Wicket
		comb.BladeFlowTPH = current + 1.0
		_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.BladeFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindBladeLoop(holder, ctx)
	defer a.cancelBladeLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Wicket
		comb.BladeFlowTPH += 0.5
		_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.BladeFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
