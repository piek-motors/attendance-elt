package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserFromCSV(t *testing.T) {
	raw := []string{"1", "John", "home", "Doe", "1234567890", "1"}
	fieldIndex := make(map[string]int)
	fieldIndex["id"] = 0
	fieldIndex["name"] = 3
	fieldIndex["lastname"] = 1
	fieldIndex["CardNo"] = 4

	t.Run("default case", func(t *testing.T) {
		// user := new(User)
		user, err := UserFromCSV(raw, fieldIndex)

		assert.Nil(t, err)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "1234567890", user.Card)

	})

	t.Run("empty card", func(t *testing.T) {
		data := make([]string, len(raw))
		copy(data, raw)
		data[fieldIndex["CardNo"]] = ""

		_, err := NewEventFromDBRecord(data, fieldIndex)

		assert.NotNil(t, err)
	})
}

func TestUserAddEvents(t *testing.T) {
	t.Run("add event & sort", func(t *testing.T) {
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
			{
				ID:        7052,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Now().AddDate(0, 1, 0),
			},
		}

		user := *new(User)
		user.AddEvents(events)

		assert.Equal(t, 3, len(user.Events))
		assert.Equal(t, 7050, user.Events[0].ID)
		assert.Equal(t, 7051, user.Events[1].ID)
		assert.Equal(t, 7052, user.Events[2].ID)
	})
}

func TestUserAdd(t *testing.T) {
	t.Run("add event & sort", func(t *testing.T) {
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
			{
				ID:        7052,
				Card:      "1213363737",
				PointName: "КПП ЦЕНТР",
				Time:      time.Now().AddDate(0, 1, 0),
			},
		}

		user := *new(User)
		user.AddEvents(events)

		assert.Equal(t, 3, len(user.Events))
		assert.Equal(t, 7050, user.Events[0].ID)
		assert.Equal(t, 7051, user.Events[1].ID)
		assert.Equal(t, 7052, user.Events[2].ID)
	})
}
