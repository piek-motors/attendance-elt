package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventFromCSV(t *testing.T) {
	raw := []string{"7050", "06/19/21 06:21:43", "62", "1213363737", "3", "DGD0340010080700164", "КПП ЦЕНТР", "КПП ЦЕНТР-1"}
	fieldIndex := make(map[string]int)
	fieldIndex["id"] = 0
	fieldIndex["card_no"] = 3
	fieldIndex["event_point_name"] = 6
	fieldIndex["time"] = 1

	t.Run("default case", func(t *testing.T) {
		event, err := NewEventFromDBRecord(raw, fieldIndex)

		assert.Nil(t, err)
		assert.Equal(t, 7050, event.ID)
		assert.Equal(t, "1213363737", event.Card)
		assert.Equal(t, "КПП ЦЕНТР", event.PointName)
		assert.Equal(t, time.Date(2021, 6, 19, 6, 21, 43, 0, time.UTC), event.Time)
	})

}

func TestFilterEvents(t *testing.T) {
	events := []Event{
		{

			ID:        7050,
			Card:      "1213363737",
			PointName: "КПП ЦЕНТР",
			Time:      time.Now().AddDate(0, -4, 0),
		},
		{
			ID:        7051,
			Card:      "1213363737",
			PointName: "КПП ЦЕНТР",
			Time:      time.Now().AddDate(0, -2, 0),
		},
	}

	t.Run("default case", func(t *testing.T) {
		events := SelectEventsForNLastMonths(events, 3)

		assert.Equal(t, 1, len(events))
		assert.Equal(t, 7051, events[0].ID)
	})
}

func TestFixEventCollision(t *testing.T) {
	t.Run("add event & sort", func(t *testing.T) {
		events := []Event{
			{
				ID:        7050,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Now().Add(time.Second * -5),
			},
			{
				ID:        7051,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Now(),
			},
			{
				ID:        7052,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Now().Add(time.Second * (EVENT_COLLISION_JITTER_SEC - 1)),
			},
		}

		goodEventIndex := CheckCollisionPresence(events, 0)
		assert.Equal(t, 2, goodEventIndex)
	})
}

func TestMartEventDirection(t *testing.T) {

	t.Run("mark event as ent", func(t *testing.T) {
		events := []Event{
			{
				ID:        7050,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Date(2021, 12, 15, 8, 27, 11, 0, time.UTC),
			},
			{
				ID:        7051,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Date(2021, 12, 15, 16, 27, 11, 0, time.UTC),
			},
			{
				ID:        7052,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Date(2021, 12, 15, 17, 27, 11, 0, time.UTC),
			},
			{
				ID:        7052,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Date(2021, 12, 18, 1, 27, 11, 0, time.UTC),
			},
			{
				ID:        7052,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Date(2021, 12, 18, 19, 27, 11, 0, time.UTC),
			},
		}

		SetEventDirection(events)

		assert.Equal(t, EventTypeEnt, events[0].Direction)
		assert.Equal(t, EventTypeExt, events[1].Direction)
		assert.Equal(t, EventTypeEnt, events[2].Direction)
		assert.Equal(t, EventTypeEnt, events[3].Direction)
		assert.Equal(t, EventTypeEnt, events[4].Direction)
	})
}

func TestConstructIntervals(t *testing.T) {
	events := []Event{
		{
			ID:        7050,
			Card:      "1213363737",
			Direction: EventTypeEnt,
			Time:      time.Date(2021, 12, 15, 8, 27, 11, 0, time.UTC),
		},
		{
			ID:        7051,
			Card:      "1213363737",
			Direction: EventTypeExt,
			Time:      time.Date(2021, 12, 15, 16, 27, 11, 0, time.UTC),
		},
		{
			ID:        7052,
			Card:      "1213363737",
			Direction: EventTypeEnt,
			Time:      time.Date(2021, 12, 15, 17, 27, 11, 0, time.UTC),
		},
		{
			ID:        7053,
			Card:      "1213363737",
			Direction: EventTypeEnt,
			Time:      time.Date(2021, 12, 18, 1, 27, 11, 0, time.UTC),
		},
		{
			ID:        7054,
			Card:      "1213363737",
			Direction: EventTypeEnt,
			Time:      time.Date(2021, 12, 18, 19, 27, 11, 0, time.UTC),
		},
	}

	t.Run("ok case", func(t *testing.T) {
		intervals := ConstructIntervals(events)

		assert.Equal(t, 3, len(intervals))
		assert.NotNil(t, intervals[0].Ent)
	})
}
