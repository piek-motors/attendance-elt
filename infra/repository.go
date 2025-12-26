package infra

import (
	"fmt"
	"log"
	"time"

	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spooky-finn/piek-attendance-prod/entity"
)

type Employee struct {
	ID        int            `db:"id"`
	FirstName string         `db:"firstname"`
	LastName  string         `db:"lastname"`
	Card      string         `db:"card"`
	CreatedAt sql.NullString `db:"created_at"`
}

type Event struct {
	ID        int       `db:"id"`
	Card      string    `db:"card"`
	Timestamp time.Time `db:"timestamp"`
}

type Interval struct {
	Ent        string         `db:"ent"`
	Ext        sql.NullString `db:"ext"`
	Card       string         `db:"card"`
	Database   string         `db:"database"`
	EntEventID int            `db:"ent_event_id"`
	ExtEventID sql.NullInt64  `db:"ext_event_id"`
}

type Repository struct {
	*sqlx.DB
}

func Connect(dataSourceName string) (*Repository, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &Repository{db}, nil
}

func (db *Repository) EmployeesAll() (employees []Employee, err error) {
	err = db.Select(&employees, "SELECT * FROM attendance.employees")
	return employees, err
}

func (db *Repository) InsertEmployees(employees []Employee) error {
	if len(employees) == 0 {
		return nil
	}
	tx := db.MustBegin()
	t := time.Now().Local().Format("2006-01-02T15:04:05")
	for _, user := range employees {
		tx.MustExec("INSERT INTO attendance.employees (firstname, lastname, card, created_at) VALUES ($1, $2, $3, $4)",
			user.FirstName, user.LastName, user.Card, t)
	}
	return tx.Commit()
}

func (db *Repository) UpdateEmployees(employees []Employee) error {
	if len(employees) == 0 {
		return nil
	}
	tx := db.MustBegin()
	for _, user := range employees {
		tx.MustExec("UPDATE attendance.employees SET firstname = $1, lastname = $2 WHERE card = $3",
			user.FirstName, user.LastName, user.Card)
	}
	return tx.Commit()
}

func (db *Repository) InsertIntervals(intervals []Interval) error {
	if len(intervals) == 0 {
		return nil
	}
	res, err := db.NamedExec(`INSERT INTO attendance.intervals (ent, ext, card, database, ent_event_id, ext_event_id)
	VALUES (:ent, :ext, :card, :database, :ent_event_id, :ext_event_id) ON CONFLICT DO NOTHING RETURNING *`, intervals)
	if err != nil {
		return fmt.Errorf("inserting intervals: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("inserting intervals: %w", err)
	}
	log.Println("inserted", ra, "intervals")
	return err

}

func (db *Repository) InsertEvents(events []entity.Event) error {
	if len(events) == 0 {
		return nil
	}
	infraEvents := make([]Event, len(events))
	for i, e := range events {
		infraEvents[i] = Event{
			ID:        e.ID,
			Card:      e.Card,
			Timestamp: e.Time,
		}
	}
	res, err := db.NamedExec(`INSERT INTO attendance.events (id, card, timestamp)
	VALUES (:id, :card, :timestamp) ON CONFLICT DO NOTHING`, infraEvents)
	if err != nil {
		return fmt.Errorf("inserting events: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("inserting events: %w", err)
	}
	log.Println("inserted", ra, "events")
	return nil
}

func (db *Repository) SyncEmployees(deviceUsers []*entity.User) error {
	existingEmployees, err := db.EmployeesAll()
	if err != nil {
		return fmt.Errorf("fail to load employees: %w", err)
	}

	insert := make([]Employee, 0)
	update := make([]Employee, 0)

	for _, deviceUser := range deviceUsers {
		var found bool
		user := Employee{
			FirstName: deviceUser.FirstName,
			LastName:  deviceUser.LastName,
			Card:      deviceUser.Card,
		}

		for _, existing := range existingEmployees {
			if user.Card == existing.Card {
				found = true

				if user.FirstName != existing.FirstName || user.LastName != existing.LastName {
					update = append(update, user)
				}

				break
			}
		}

		if !found {
			insert = append(insert, user)
		}
	}

	log.Printf("inserted %d employees\n", len(insert))
	log.Printf("updated %d employees\n", len(update))
	err = db.UpdateEmployees(update)
	if err != nil {
		return err
	}
	return db.InsertEmployees(insert)
}
