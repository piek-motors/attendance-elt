package entity

import (
	"fmt"
	"strconv"
	"time"
)

const IDEAL_WORKSHIFT_DUR = 14
const EVENT_COLLISION_JITTER_SEC = 300

type Direction string

const (
	EventTypeEnt Direction = "ent"
	EventTypeExt Direction = "ext"
)

type Event struct {
	ID        int
	Card      string
	PointName string
	Time      time.Time
	Direction Direction
}

func NewEventFromDBRecord(record []string, index map[string]int) (Event, error) {
	e := Event{}

	id, err := strconv.Atoi(record[index["id"]])
	if err != nil {
		return Event{}, fmt.Errorf("error parsing id: %w", err)
	}

	e.ID = id
	e.Card = record[index["card_no"]]
	e.PointName = record[index["event_point_name"]]
	e.Time, err = time.Parse("01/02/06 15:04:05", record[index["time"]])

	if err != nil {
		return Event{}, fmt.Errorf("NewEventFromDBRecord: %w", err)
	}

	return e, nil
}

func (e *Event) IsValid() bool {
	if e.Card == "" || e.PointName == "" || e.Time.IsZero() {
		return false
	}

	return true
}

func SelectEventsForNLastMonths(events []Event, n int) []Event {
	var result []Event
	now := time.Now()

	for _, event := range events {
		if event.Time.After(now.AddDate(0, -n, 0)) {
			result = append(result, event)
		}
	}

	return result
}

/*
 * Marking events by type (ent or ext)
 * based on the module distance to the previous event
 */
func SetEventDirection(events []Event) {
	for i := range events {
		if i+1 >= len(events) {
			break
		}

		cur := &events[i]
		nextEvent := &events[i+1]
		timedelta := nextEvent.Time.Sub(cur.Time).Hours()

		// DANGER: don't send the first event because it doesn't reflect the real direction
		if i == 0 {
			events[i].Direction = EventTypeEnt
		}
		if timedelta < IDEAL_WORKSHIFT_DUR && cur.Direction == EventTypeEnt {
			nextEvent.Direction = EventTypeExt
		} else {
			nextEvent.Direction = EventTypeEnt
		}
	}
}

/*
 * Anthicollision algorithm, descends into primary sampling algorithm
 * The essence of the algorithm is that if a person has 2 events at +- the same time
 * then the algorithm recursively skips their similar ones so as not to spoil the statistics
 */
func ExcludeCollisions(events []Event) []Event {
	result := make([]Event, 0, len(events))

	for i := 0; i < len(events); {
		goodEventIndex := CheckCollisionPresence(events, i)

		result = append(result, events[goodEventIndex])
		i = goodEventIndex + 1
	}

	return result
}

func CheckCollisionPresence(events []Event, i int) int {
	if i+1 >= len(events) {
		return i
	}

	cur := events[i]
	next := events[i+1]
	timedelta := next.Time.Sub(cur.Time)

	if timedelta.Seconds() < EVENT_COLLISION_JITTER_SEC {
		return CheckCollisionPresence(events, i+1)
	}

	return i
}
