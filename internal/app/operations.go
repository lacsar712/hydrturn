package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/hydrturn/internal/fsm"
	"github.com/lacsar712/hydrturn/internal/interlock"
	"github.com/lacsar712/hydrturn/internal/model"
)

func (a *App) StartSync(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.interlock.Gate().Allow(a.cfg.UnitID, snap.State); err != nil {
			return err
		}
		comb, err := a.wicket.StartSync(ctx, snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvStartSync)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("start_sync", storePayload("holder", holder))
		return nil
	})
}

func (a *App) CompleteSync(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.wicket.CompleteSync(snap.Wicket); err != nil {
			return err
		}
		state, err := a.fsm.Dispatch(ctx, fsm.EvSyncComplete)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("sync_complete", "")
		return nil
	})
}

func (a *App) Ignite(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		a.refreshPermissives(snap)
		if err := a.permissives.CheckIgnition(); err != nil {
			return err
		}
		comb, err := a.wicket.Ignite(ctx, snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvIgnite)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ignite", storePayload("holder", holder))
		return nil
	})
}

func (a *App) StabilizeIgnition(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		comb, err := a.wicket.Stabilize(snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvIgnitionStable)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ignition_stable", "")
		return nil
	})
}

func (a *App) RampLoad(ctx context.Context, holder string, loadPct float64) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.interlock.AuthorizeFiring(a.cfg.UnitID, holder, snap.State, now); err != nil {
			return err
		}
		comb := a.wicket.RampToLoad(snap, loadPct)
		_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
		turbinelnReading, err := a.turbineln.Tick(ctx, snap, true)
		if err != nil {
			return err
		}
		_ = a.store.UpdateTurbineln(a.cfg.UnitID, turbinelnReading)
		state, err := a.fsm.Dispatch(ctx, fsm.EvReachFiring)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ramp_load", fmt.Sprintf(`{"load_pct":%f}`, loadPct))
		return nil
	})
}

func (a *App) Trip(ctx context.Context, reason string) error {
	snap := a.Snapshot()
	comb := a.wicket.Trip(snap.Wicket)
	_ = a.store.UpdateWicket(a.cfg.UnitID, comb)
	a.interlock.Gate().Block(a.cfg.UnitID, "trip")
	state, err := a.fsm.Dispatch(ctx, fsm.EvTrip)
	if err != nil {
		return err
	}
	a.syncState(state)
	a.journalEvent("trip", storePayload("reason", reason))
	return nil
}

func (a *App) ResetTrip(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		a.interlock.Gate().Unblock(a.cfg.UnitID)
		state, err := a.fsm.Dispatch(ctx, fsm.EvResetTrip)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("reset_trip", "")
		return nil
	})
}

func (a *App) UpdateSettings(ctx context.Context, settings model.PlantSettings) error {
	select {
	case <-ctx.Done():
		return model.ErrContextDone
	default:
	}
	if err := a.turbineln.ValidateSettings(settings); err != nil {
		return err
	}
	return a.store.UpdateSettings(a.cfg.UnitID, settings)
}

func (a *App) EnterService(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		state, err := a.fsm.Dispatch(ctx, fsm.EvEnterService)
		if err != nil {
			return err
		}
		a.syncState(state)
		return nil
	})
}

func (a *App) OnSpinLoss(ctx context.Context, holder string) error {
	_ = holder
	snap := a.Snapshot()
	if err := interlock.CheckSpinLoss(snap.Wicket); err != nil {
		return fmt.Errorf("spin loss: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		a.StopTickLoop()
		a.cancelBladeLoop(holder)
		state, err := a.fsm.Dispatch(ctx, fsm.EvShutdown)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("shutdown", "")
		return nil
	})
}

func (a *App) StartTickLoop(parent context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.tickCancels[a.tickLabel()]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	a.tickCancels[a.tickLabel()] = cancel
	interval := a.cfg.TickInterval
	if interval <= 0 {
		interval = time.Second
	}
	_ = a.scheduler.Schedule(ctx, a.tickLabel(), interval, a.runTick)
}

func (a *App) StopTickLoop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.tickCancels[a.tickLabel()]; ok {
		cancel()
		delete(a.tickCancels, a.tickLabel())
	}
}

func (a *App) runTick(ctx context.Context) error {
	snap := a.Snapshot()
	firing := a.isFiring(snap.State)
	draftubeReading, err := a.draftube.Tick(ctx, snap, firing)
	if err != nil {
		return err
	}
	_ = a.store.UpdateDraftube(a.cfg.UnitID, draftubeReading)
	snap, _ = a.store.Require(a.cfg.UnitID)
	turbinelnReading, err := a.turbineln.Tick(ctx, snap, firing)
	if err != nil {
		return err
	}
	_ = a.store.UpdateTurbineln(a.cfg.UnitID, turbinelnReading)
	a.refreshPermissives(a.Snapshot())
	a.telemetry.RecordTick(firing)
	if firing && a.turbineln.Pressure().TripRequired(turbinelnReading.SteamPressurePSI) {
		_ = a.Trip(ctx, "steam_pressure")
		return nil
	}
	if a.draftube.TripRequired(draftubeReading) {
		_ = a.Trip(ctx, "draftube_level")
	}
	return nil
}
