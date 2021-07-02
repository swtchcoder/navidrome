package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upAddUserPrefsPlayerScrobblerEnabled, downAddUserPrefsPlayerScrobblerEnabled)
}

func upAddUserPrefsPlayerScrobblerEnabled(tx *sql.Tx) error {
	err := upAddUserPrefs(tx)
	if err != nil {
		return err
	}
	return upPlayerScrobblerEnabled(tx)
}

func upAddUserPrefs(tx *sql.Tx) error {
	_, err := tx.Exec(`
create table user_props
(
    user_id varchar not null,
    key     varchar not null,
    value   varchar,
    constraint user_props_pk
        primary key (user_id, key)
);
`)
	return err
}

func upPlayerScrobblerEnabled(tx *sql.Tx) error {
	_, err := tx.Exec(`
alter table player add scrobble_enabled bool default true;
`)
	return err
}

func downAddUserPrefsPlayerScrobblerEnabled(tx *sql.Tx) error {
	return nil
}
