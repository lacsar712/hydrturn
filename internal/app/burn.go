package app

import (
	"context"
	"time"
)

func (a *App) RunWarmupBurnScheduler(ctx context.Context, ignitionAt time.Time) error {
	for !a.warmupWindow.Ready(ignitionAt) {
		if err := ctx.Err(); err != nil {
			return err
		}
		a.advanceClock(100 * time.Millisecond)
	}
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return err
	}
	// Thread the operator's context through so a warmup-window withdrawal
	// (ctx cancellation) is honored at the plan layer, not just the HMI/journal layer.
	// Passing context.Background() here would decouple plan installation from the
	// operator's cancel and leave the old blade ramp-up steps (PlanBurnSchedule)
	// appended to the day's plan after the window was withdrawn.
	return a.scheduler.InstallBurnPlanCtx(ctx, snap.Settings, "warmup-burn")
}

func (a *App) SchedulerItemCount() int {
	return a.scheduler.ItemCount()
}
