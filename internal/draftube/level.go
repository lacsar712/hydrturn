package draftube

import (
	"math"

	"github.com/lacsar712/hydrturn/internal/clock"
	"github.com/lacsar712/hydrturn/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.DraftubeCondition) {
	level := snap.Draftube.LevelPercent
	if !firing {
		return level, model.DraftubeNormal
	}
	balance := snap.Draftube.FeedwaterTPH - snap.Draftube.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinDraftubeLevelPercent, math.Min(model.MaxDraftubeLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.DraftubeCondition {
	setpoint := snap.Settings.DraftubeLevelSetpoint
	if level > setpoint+15 {
		return model.DraftubeSwell
	}
	if level < setpoint-15 {
		return model.DraftubeShrink
	}
	if snap.Turbineln.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.DraftubeCarry
	}
	return model.DraftubeNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.DraftubeLevelSetpoint - snap.Draftube.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinDraftubeLevelPercent && level <= model.MaxDraftubeLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripDraftubeLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripDraftubeHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Draftube.LevelPercent - snap.Settings.DraftubeLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Draftube.SteamFlowTPH
	feed := snap.Draftube.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
