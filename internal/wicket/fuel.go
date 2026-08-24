package wicket

import (
	"math"

	"github.com/lacsar712/hydrturn/internal/clock"
	"github.com/lacsar712/hydrturn/internal/model"
)

type BladeRegulator struct {
	clk clock.ProcessClock
}

func NewBladeRegulator(clk clock.ProcessClock) *BladeRegulator {
	return &BladeRegulator{clk: clk}
}

func (f *BladeRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.BladeFlowTPH * 0.08
}

func (f *BladeRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.BladeFlowTPH * loadPct
}

func (f *BladeRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *BladeRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *BladeRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *BladeRegulator) ValidatePermissive(settings model.PlantSettings, draftubeOK, syncOK bool) error {
	if !syncOK {
		return model.ErrSyncIncomplete
	}
	if !draftubeOK {
		return model.ErrDraftubeLevelTrip
	}
	if settings.BladeFlowTPH <= 0 {
		return model.ErrBladePermissive
	}
	return nil
}

func (f *BladeRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.BladeFlowTPH * 0.2
}

func (f *BladeRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.BladeFlowTPH * 1.1
}
