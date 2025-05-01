package entity

import (
	"fmt"
	"time"
)

type Interval struct {
	Ent *Event
	Ext *Event
}

func (i *Interval) Dur() time.Duration {
	if i.Ext == nil {
		return 0
	}
	return i.Ext.Time.Sub(i.Ent.Time)
}

func (i *Interval) String() string {
	ext := "nil"
	if i.Ext != nil {
		ext = i.Ext.Time.Format("2006-01-02T15:04:05")
	}

	return fmt.Sprintf("ent: %v, ext: %v, dur: %s", i.Ent.Time.Format(time.RFC3339), ext, i.Dur())
}

func ConstructIntervals(events []Event) []Interval {
	result := make([]Interval, 0, len(events)/2)

	for i := range events {
		if i+1 >= len(events) {
			break
		}

		event := &events[i]
		nextEvent := &events[i+1]

		if event.Direction == "" || nextEvent.Direction == "" {
			panic("event direction is not marked")
		}

		if event.Direction == EventTypeEnt && nextEvent.Direction == EventTypeExt {
			result = append(result, Interval{
				Ent: event,
				Ext: nextEvent,
			})
		} else if event.Direction == EventTypeEnt && nextEvent.Direction == EventTypeEnt {
			result = append(result, Interval{
				Ent: event,
				Ext: nil,
			})
		}
	}

	return result
}
