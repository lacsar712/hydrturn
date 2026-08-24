package api

import (
	"errors"

	"github.com/lacsar712/hydrturn/internal/model"
)

func classifyDraftubeError(err error) (string, bool) {
	if errors.Is(err, model.ErrDraftubeLevelLow) {
		return "draftube_level_low", true
	}
	return "", false
}
