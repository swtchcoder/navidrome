package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up20200512104202, Down20200512104202)
}

func Up20200512104202(tx *sql.Tx) error {
	_, err := tx.Exec(`
alter table media_file 
    add disc_subtitle varchar(255);
    `)
	if err != nil {
		return err
	}
	notice(tx, "A full rescan will be performed to import disc subtitles")
	return forceFullRescan(tx)
}

func Down20200512104202(tx *sql.Tx) error {
	return nil
}
