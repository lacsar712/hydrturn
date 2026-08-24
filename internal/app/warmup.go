package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/hydrturn/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Wicket.SyncStartedAt.IsZero() {
		return false, "sync not started"
	}
	if !a.syncWindow.Ready(snap.Wicket.SyncStartedAt) {
		return false, "sync window open"
	}
	if !snap.Wicket.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Wicket.IgnitionAt) {
		return false, "wicket warmup window open"
	}
	if !snap.Draftube.LastSwellAt.IsZero() {
		if err := a.draftube.RequireSettled(snap.Draftube); err != nil {
			return false, "draftube swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) SyncRemaining() string {
	snap := a.Snapshot()
	if snap.Wicket.SyncStartedAt.IsZero() {
		return "not started"
	}
	if a.syncWindow.Ready(snap.Wicket.SyncStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) WicketWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Wicket.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Wicket.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
