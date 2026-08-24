package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			DraftubeLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			BladeFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Draftube: DraftubeReading{
			LevelPercent: 50,
			Condition:    DraftubeNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Wicket: WicketReading{
			BurnerPhase: BurnerIdle,
		},
		Turbineln: TurbinelnReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) DraftubeWithinLimits() bool {
	return s.Draftube.LevelPercent >= MinDraftubeLevelPercent && s.Draftube.LevelPercent <= MaxDraftubeLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Turbineln.SteamPressurePSI <= MaxSteamPressurePSI
}
