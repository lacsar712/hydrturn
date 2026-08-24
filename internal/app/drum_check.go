package app

import (
	"fmt"

	"github.com/lacsar712/hydrturn/internal/model"
)

func (a *App) CheckDraftubeLevel(snap model.PlantSnapshot) error {
	if snap.Draftube.LevelPercent < model.MinDraftubeLevelPercent {
		return fmt.Errorf("%w", model.ErrDraftubeLevelLow)
	}
	return nil
}
