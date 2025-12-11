package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	entity "github.com/spooky-finn/piek-attendance-prod/entity"
	"github.com/spooky-finn/piek-attendance-prod/infra"

	database "github.com/spooky-finn/piek-attendance-prod/infra"
)

var (
	selectEventsForMonths = flag.Int("selectfor", 2, "select events for last n months")
)

func main() {
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	mdbpath := os.Getenv("ACCESS_MDB_PATH")
	exporter := infra.NewMdbExporter(mdbpath)

	users, err := exporter.ExportUsersFromDB()
	if err != nil {
		log.Fatalln(err)
	}

	events, err := exporter.ExportEventsFromDB(*selectEventsForMonths)
	if err != nil {
		log.Fatalf("error exporting events: %v", err)
	}

	destDBconnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	db, err := database.Connect(destDBconnStr)
	log.Println("db connection established")
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	err = db.SyncEmployees(users)
	if err != nil {
		log.Fatalf("error syncing users: %v", err)
	}

	eventsmap := make(map[string][]entity.Event)
	for _, event := range events {
		eventsmap[event.Card] = append(eventsmap[event.Card], event)
	}

	intervals := make([]infra.Interval, 0)
	for _, user := range users {
		user.AddEvents(eventsmap[user.Card])
		user.RunFlow(*selectEventsForMonths)

		for _, interval := range user.Intervals {
			extTime := "nil"
			extId := 0
			if interval.Ext != nil {
				extTime = interval.Ext.Time.Format("2006-01-02T15:04:05")
				extId = interval.Ext.ID
			}

			intervals = append(intervals, infra.Interval{
				Ent:        interval.Ent.Time.Format("2006-01-02T15:04:05"),
				Card:       user.Card,
				Ext:        sql.NullString{String: extTime, Valid: extTime != "nil"},
				Database:   os.Getenv("CONTROLLER_DIVISION_NAME"),
				EntEventID: interval.Ent.ID,
				ExtEventID: sql.NullInt64{
					Int64: int64(extId),
					Valid: extId != 0,
				},
			})
		}
	}
	log.Printf("formed %v intervals for last %v months \n", len(intervals), *selectEventsForMonths)

	err = db.InsertIntervals(intervals)
	if err != nil {
		log.Fatalf("error getting intervals: %v", err)
	}
}
