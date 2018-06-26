package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00002, Down00002)
}

// Up00002 adds a new column to the scenes table to contain custom bounding boxes
func Up00002(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE public.scenes ADD COLUMN bbox json;`)
	return err
}

// Down00002 undoes the effects of Up00002
func Down00002(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE public.scenes DROP COLUMN bbox;`)
	return err
}
