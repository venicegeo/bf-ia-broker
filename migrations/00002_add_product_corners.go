package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00002, Down00002)
}

//Up00002 adds the columns for the product corner points.
func Up00002(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		ALTER TABLE public.scenes ADD COLUMN IF NOT EXISTS corner_ul geometry;
		ALTER TABLE public.scenes ADD COLUMN IF NOT EXISTS corner_ur geometry;
		ALTER TABLE public.scenes ADD COLUMN IF NOT EXISTS corner_ll geometry;
		ALTER TABLE public.scenes ADD COLUMN IF NOT EXISTS corner_lr geometry;
		`)
	return err
}

//Down00002 removes the columns.
func Down00002(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`
		ALTER TABLE public.scenes DROP COLUMN IF  EXISTS corner_ul ;
		ALTER TABLE public.scenes DROP COLUMN IF  EXISTS corner_ur ;
		ALTER TABLE public.scenes DROP COLUMN IF  EXISTS corner_ll ;
		ALTER TABLE public.scenes DROP COLUMN IF  EXISTS corner_lr ;
		`)
	return err
}
