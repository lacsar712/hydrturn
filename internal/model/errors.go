package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrBladePermissive   = errors.New("blade permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrDraftubeLevelTrip    = errors.New("draftube level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrWicketTrip   = errors.New("wicket trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrSyncIncomplete  = errors.New("hublock sync incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrDraftubeLevelLow     = errors.New("draftube level below low limit")
	ErrSpinLoss        = errors.New("hublock spin lost")
	ErrReliefLimit    = errors.New("relief valve at limit")
)
