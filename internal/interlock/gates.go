package interlock

import (
	"fmt"

	"github.com/lacsar712/hydrturn/internal/model"
)

type PermissiveSet struct {
	bladeOK       bool
	ignitionOK   bool
	draftubeOK       bool
	pressureOK   bool
	wicketOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetBlade(ok bool)       { p.bladeOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetDraftube(ok bool)       { p.draftubeOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetWicket(ok bool) { p.wicketOK = ok }

func (p *PermissiveSet) BladeOK() bool       { return p.bladeOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) DraftubeOK() bool       { return p.draftubeOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) WicketOK() bool { return p.wicketOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.bladeOK && p.ignitionOK && p.draftubeOK && p.pressureOK && p.wicketOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.bladeOK {
		return fmt.Errorf("%w", model.ErrBladePermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckSpinLoss(reading model.WicketReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.HublockTempF < 600 {
		return fmt.Errorf("%w", model.ErrSpinLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.draftubeOK {
		return fmt.Errorf("%w", model.ErrDraftubeLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.wicketOK {
		return fmt.Errorf("%w", model.ErrWicketTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
