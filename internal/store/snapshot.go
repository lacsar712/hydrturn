package store

import "github.com/lacsar712/hydrturn/internal/model"

type DraftubeSnapshotView struct {
	UnitID   string
	Draftube     model.DraftubeReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneDraftubeSnapshot(s model.PlantSnapshot) DraftubeSnapshotView {
	out := DraftubeSnapshotView{
		UnitID:   s.UnitID,
		Draftube:     s.Draftube,
		Revision: s.Revision,
	}
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
