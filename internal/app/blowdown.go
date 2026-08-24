package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/hydrturn/internal/model"
)

const maxReliefOpeningPct = 100.0

func (a *App) OpenRelief(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxReliefOpeningPct {
		return fmt.Errorf("relief: %w", model.ErrReliefLimit)
	}
	return nil
}

func (a *App) ReliefAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxReliefOpeningPct {
		return fmt.Errorf("unknown fault")
	}
	return nil
}
