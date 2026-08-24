package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	SyncWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	DraftubeSwellSettleWindow  = 45 * time.Second
	WicketWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxDraftubeLevelPercent    = 95.0
	MinDraftubeLevelPercent    = 15.0
	TripDraftubeLowPercent     = 10.0
	TripDraftubeHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinHublockO2Percent    = 2.5
	MaxHublockO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
