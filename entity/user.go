package entity

import (
	"fmt"
	"sort"
)

type User struct {
	FirstName string
	LastName  string
	Card      string
	Events    []Event
	Intervals []Interval
}

func UserFromCSV(record []string, index map[string]int) (*User, error) {
	u := &User{}
	u.FirstName = record[index["lastname"]]
	u.LastName = record[index["name"]]
	u.Card = record[index["CardNo"]]
	u.Intervals = make([]Interval, 0)

	if u.Card == "" {
		return &User{}, fmt.Errorf("card number is empty for user: %s %s", u.FirstName, u.LastName)
	}
	return u, nil
}

func (u *User) AddEvents(ev []Event) {
	u.Events = make([]Event, 0)

	for _, event := range ev {
		if !event.IsValid() {
			continue
		}
		u.Events = append(u.Events, event)
	}

	sort.Slice(u.Events, func(i, j int) bool {
		return u.Events[i].Time.Before(u.Events[j].Time)
	})
}

func (u *User) RunFlow(selectEventsFor int) {
	res := ExcludeCollisions(u.Events)
	SetEventDirection(res)

	u.Events = SelectEventsForNLastMonths(res, selectEventsFor)
	u.Intervals = ConstructIntervals(res)
}
