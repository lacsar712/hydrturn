package wicket

import (
	"math"

	"github.com/lacsar712/hydrturn/internal/clock"
	"github.com/lacsar712/hydrturn/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateHublockTemp(reading model.WicketReading) float64 {
	base := 300.0
	bladeHeat := reading.BladeFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + bladeHeat - airCool
}

func (b *BurnerController) SpinStable(reading model.WicketReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.HublockTempF > 800 && reading.ExcessO2Pct >= model.MinHublockO2Percent
}

func (b *BurnerController) TripRequired(reading model.WicketReading) bool {
	if reading.ExcessO2Pct > model.MaxHublockO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.HublockTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerSync:
		return "Sync"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Spin"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.WicketReading) float64 {
	return reading.BladeFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentBlade float64) float64 {
	if settings.BladeFlowTPH <= 0 {
		return 0
	}
	return currentBlade / settings.BladeFlowTPH
}

func (b *BurnerController) MinStableBlade(settings model.PlantSettings) float64 {
	return settings.BladeFlowTPH * 0.25
}

func (b *BurnerController) NormalizeBlade(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
