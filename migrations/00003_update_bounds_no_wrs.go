package migration

import (
	"database/sql"
	"errors"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00003, Down00003)
}

//Up00003 updates scenes.bounds to not be based on WRS bounding box
//this process is destructive and cannot be undone
func Up00003(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
	UPDATE scenes SET bounds = q1.safePoly FROM (
		SELECT product_id AS pid, (
			CASE WHEN (
				(abs(st_x(corner_ul)-st_x(corner_ur)) > 90) OR 
				(abs(st_x(corner_ul)-st_x(corner_lr))>90) OR  
				(abs(st_x(corner_ul)-st_x(corner_ll))>90)
				) THEN (
					st_union(
						st_intersection(
							st_makeenvelope(-180, -90, 180, 90, 4326),
							st_makepolygon(st_makeline(array[st_wrapx(corner_ul, 0, 360), st_wrapx(corner_ur, 0, 360), st_wrapx(corner_lr, 0, 360), st_wrapx(corner_ll, 0, 360), st_wrapx(corner_ul, 0, 360)]))
						),
						st_intersection(
							st_makeenvelope(-180, -90, 180, 90, 4326),
							st_makepolygon(st_makeline(array[st_wrapx(corner_ul, 0, -360), st_wrapx(corner_ur, 0, -360), st_wrapx(corner_lr, 0, -360), st_wrapx(corner_ll, 0, -360), st_wrapx(corner_ul, 0, -360)]))
						)
					)
				) 
			ELSE bounds END
		) AS safePoly
		FROM scenes 
		WHERE corner_ll IS NOT NULL
	) AS q1 
	WHERE product_id = q1.pid;
	`)
	return err
}

//Down00003 would undo the migration, but this downgrade is not supported
func Down00003(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return errors.New("Up00003 was destructive and cannot be rolled back")
}
