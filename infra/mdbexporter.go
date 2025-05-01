package infra

import (
	"bytes"
	"log"
	"os/exec"
	"runtime"

	"github.com/spooky-finn/piek-attendance-prod/entity"
	"golang.org/x/text/encoding/charmap"
)

type MdbExporter struct {
	dblocation  string
	mdbToolsBin string
}

func NewMdbExporter(mdbpath string) *MdbExporter {
	mdbToolsBin := "mdb-export"

	if runtime.GOOS == "windows" {
		mdbToolsBin = "./mdbtools-win/mdb-export"
	}

	return &MdbExporter{dblocation: mdbpath, mdbToolsBin: mdbToolsBin}
}

func (e *MdbExporter) ExportEventsFromDB(selectFor int) ([]entity.Event, error) {
	out, errout, err := e.mdbExport(e.dblocation, "acc_monitor_log")

	if err != nil {
		log.Fatalln("err: exec: ", errout, err)
		return nil, err
	}

	events, err := SerializeCSVInput(out, entity.NewEventFromDBRecord)
	events = entity.SelectEventsForNLastMonths(events, selectFor+1)

	if err != nil {
		log.Fatalln("err", err)
	}

	return events, nil
}

func (e *MdbExporter) ExportUsersFromDB() ([]*entity.User, error) {
	out, errout, err := e.mdbExport(e.dblocation, "USERINFO")

	if err != nil {
		log.Fatalln("err: exec: ", errout)
		return nil, err
	}

	events, err := SerializeCSVInput(out, entity.UserFromCSV)

	if err != nil {
		log.Fatalln("err", err)
	}

	return events, nil
}

func (e *MdbExporter) mdbExport(command ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(e.mdbToolsBin, command...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if runtime.GOOS == "windows" {
		// reencode stdout to UTF-8 from windows-1251
		stdout = *bytes.NewBuffer(DecodeWindows1251(cmd.Stdout.(*bytes.Buffer).Bytes()))
	}

	return stdout.String(), stderr.String(), err
}

func DecodeWindows1251(ba []uint8) []uint8 {
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(ba)
	return out
}
