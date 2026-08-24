package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/hydrturn/internal/model"
)

type TurbinelnFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	bladePermissive bool
	syncComplete  bool
	hooks          *HookChain
}

func NewTurbinelnFSM(unitID string) *TurbinelnFSM {
	_ = unitID
	return &TurbinelnFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *TurbinelnFSM) Hooks() *HookChain { return f.hooks }

func (f *TurbinelnFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *TurbinelnFSM) SetBladePermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bladePermissive = ok
}

func (f *TurbinelnFSM) SetSyncComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.syncComplete = ok
}

func (f *TurbinelnFSM) BladePermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.bladePermissive
}

func (f *TurbinelnFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		if f.hooks != nil {
			_ = f.hooks.RunAfter(ctx, f.state, f.state, event)
		}
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.bladePermissive {
		return f.state, fmt.Errorf("%w", model.ErrBladePermissive)
	}
	if event == EvSyncComplete && !f.syncComplete {
		return f.state, fmt.Errorf("%w", model.ErrSyncIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *TurbinelnFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
