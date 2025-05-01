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

type User struct {
	ID        int            `db:"id"`
	FirstName string         `db:"firstname"`
	LastName  string         `db:"lastname"`
	Card      string         `db:"card"`
	CreatedAt sql.NullString `db:"created_at"`
}

type Interval struct {
	Ent        string         `db:"ent"`
	Ext        sql.NullString `db:"ext"`
	Card       string         `db:"card"`
	Database   string         `db:"database"`
	EntEventID int            `db:"ent_event_id"`
	ExtEventID sql.NullInt64  `db:"ext_event_id"`
}

type DestDB struct {
	*sqlx.DB
}

func Connect(dataSourceName string) (*DestDB, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DestDB{db}, nil
}

func (db *DestDB) UsersAll() (users []User, err error) {
	err = db.Select(&users, "SELECT * FROM attendance.users")
	return users, err
}

func (db *DestDB) InsertUsers(users []User) error {
	if len(users) == 0 {
		return nil
	}

	tx := db.MustBegin()
	t := time.Now().Local().Format("2006-01-02T15:04:05")
	for _, user := range users {
		tx.MustExec("INSERT INTO attendance.users (firstname, lastname, card, created_at) VALUES ($1, $2, $3, $4)",
			user.FirstName, user.LastName, user.Card, t)
	}
	return tx.Commit()
}

func (db *DestDB) InsertIntervals(intervals []Interval) error {
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

func (db *DestDB) SyncUsers(deviceUsers []*entity.User) error {
	dbUsers, err := db.UsersAll()
	if err != nil {
		return fmt.Errorf("sinchronizing users: %w", err)
	}

	unregisteredUsers := make([]User, 0)

	for _, deviceUser := range deviceUsers {
		var found bool

		for _, dbUser := range dbUsers {
			if deviceUser.Card == dbUser.Card {
				found = true
				break
			}
		}

		if !found {
			unregisteredUsers = append(unregisteredUsers, User{
				FirstName: deviceUser.FirstName,
				LastName:  deviceUser.LastName,
				Card:      deviceUser.Card,
			})
		}
	}

	log.Printf("syncing %d users\n", len(unregisteredUsers))
	return db.InsertUsers(unregisteredUsers)
}
