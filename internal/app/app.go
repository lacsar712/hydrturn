package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/hydrturn/internal/turbineln"
	"github.com/lacsar712/hydrturn/internal/clock"
	"github.com/lacsar712/hydrturn/internal/wicket"
	"github.com/lacsar712/hydrturn/internal/config"
	"github.com/lacsar712/hydrturn/internal/draftube"
	"github.com/lacsar712/hydrturn/internal/fsm"
	"github.com/lacsar712/hydrturn/internal/interlock"
	"github.com/lacsar712/hydrturn/internal/model"
	"github.com/lacsar712/hydrturn/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.TurbinelnFSM
	turbineln        *turbineln.Controller
	wicket    *wicket.Coordinator
	draftube          *draftube.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	syncWindow   *clock.SyncWindow
	warmupWindow  *clock.WicketWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	bladeLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewTurbinelnFSM(cfg.UnitID),
		turbineln:       turbineln.NewController(clk),
		wicket:   wicket.NewCoordinator(clk),
		draftube:         draftube.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		syncWindow:  clock.NewSyncWindow(clk),
		warmupWindow: clock.NewWicketWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		bladeLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.TurbinelnFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetDraftube(a.draftube.Level().WithinLimits(snap.Draftube.LevelPercent))
	a.permissives.SetPressure(a.turbineln.Pressure().WithinTripLimits(snap.Turbineln.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetWicket(a.wicket.Burner().SpinStable(snap.Wicket))
	a.permissives.SetBlade(snap.Wicket.BladeFlowTPH > 0 || snap.State == model.StateSync)
	a.permissives.SetIgnition(snap.Wicket.BurnerPhase == model.BurnerStable || snap.Wicket.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetBladePermissive(a.permissives.BladeOK())
	a.fsm.SetSyncComplete(a.syncWindow.Ready(snap.Wicket.SyncStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
