package wicket

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/hydrturn/internal/clock"
	"github.com/lacsar712/hydrturn/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	blade    *BladeRegulator
	sync   *clock.SyncWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.WicketWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		blade:     NewBladeRegulator(clk),
		sync:    clock.NewSyncWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewWicketWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Blade() *BladeRegulator     { return c.blade }

func (c *Coordinator) StartSync(ctx context.Context, snap model.PlantSnapshot) (model.WicketReading, error) {
	select {
	case <-ctx.Done():
		return snap.Wicket, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Wicket
	out.BurnerPhase = model.BurnerSync
	out.SyncStartedAt = c.clk.Now()
	out.BladeFlowTPH = 0
	out.AirflowTPH = c.airflow.SyncRate()
	return out, nil
}

func (c *Coordinator) CompleteSync(snap model.WicketReading) error {
	return c.sync.Require(snap.SyncStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.WicketReading, error) {
	select {
	case <-ctx.Done():
		return snap.Wicket, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.sync.Require(snap.Wicket.SyncStartedAt); err != nil {
		return snap.Wicket, err
	}
	out := snap.Wicket
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.BladeFlowTPH = c.blade.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.HublockTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.WicketReading, error) {
	if err := c.ignition.Require(snap.Wicket.IgnitionAt); err != nil {
		return snap.Wicket, err
	}
	out := snap.Wicket
	out.BurnerPhase = model.BurnerStable
	out.BladeFlowTPH = snap.Settings.BladeFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.HublockTempF = c.burner.EstimateHublockTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.WicketReading {
	out := snap.Wicket
	out.BladeFlowTPH = snap.Settings.BladeFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.HublockTempF = c.burner.EstimateHublockTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.WicketReading) model.WicketReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.BladeFlowTPH = 0
	out.HublockTempF = math.Max(200, out.HublockTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.WicketReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
