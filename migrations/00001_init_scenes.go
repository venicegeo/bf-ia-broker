package migration

import (
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"strings"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

//Up00001 adds the wrs2 path/row related stuff
func Up00001(tx *sql.Tx) error {
	// This code is executed when the migration is applied.

	err := addTablesAndFunctions(tx)

	if err == nil {
		err = populateWRSPaths(tx)
	}

	if err == nil {
		err = addIndexes(tx)
	}

	//TODO: Add the constraints, indexes, and vacuum?

	return err
}

//Down00001 undoes the db changes.
func Down00001(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	err := dropTablesAndFunctions(tx)
	return err
}

func addIndexes(tx *sql.Tx) error {

	_, err := tx.Exec(`ALTER TABLE public.wrs2paths
		ADD CONSTRAINT wrs2paths_primary_path_row PRIMARY KEY (path, "row")
		WITH (FILLFACTOR=100);

		CREATE INDEX idx_scenes_bounds
		ON public.scenes USING gist
		(bounds);
		`)

	return err
}

func dropTablesAndFunctions(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF EXISTS public.scenes
		DROP TABLE IF EXISTS public.wrs2paths;
		DROP FUNCTION IF EXISTS public.ls_clampedtoworldboundary(geometry);
		DROP FUNCTION IF EXISTS public.longtransactionsenabled();
		`)
	return err
}

func addTablesAndFunctions(tx *sql.Tx) error {

	_, err := tx.Exec(`
		--ls_worldboundary--------------------------------------
		CREATE OR REPLACE FUNCTION public.ls_worldboundary(
		)
		RETURNS geometry
		LANGUAGE 'sql'

		COST 100
		IMMUTABLE
	AS $BODY$

	SELECT st_geomfromtext('POLYGON((-180 -90, -180 90, 180 90, 180 -90, -180 -90))', 4326);

	$BODY$;

	--ls_clampedtoworldboundary----------------------------------------
	CREATE OR REPLACE FUNCTION public.ls_clampedtoworldboundary(
		geom geometry)
		RETURNS geometry
		LANGUAGE 'sql'

		COST 100
		VOLATILE
	AS $BODY$

	SELECT
	CASE
		WHEN st_xmax(geom) - st_xmin(geom) > 180.0 THEN
			st_union(
			st_intersection(st_shiftlongitude(geom), ls_worldboundary()),
			st_intersection(st_translate(st_shiftlongitude(geom), -360.0, 0.0), ls_worldboundary())
			)
		ELSE
			geom
		END

	$BODY$;

	--wrs2paths------------------
	CREATE TABLE public.wrs2paths
(
    path smallint NOT NULL,
    "row" smallint NOT NULL,
    center geometry NOT NULL,
    boundary geometry NOT NULL
	)
	WITH (
		OIDS = FALSE
	);

	CREATE TABLE public.scenes
	(
		product_id text COLLATE pg_catalog."default" NOT NULL,
		acquisition_date timestamp without time zone NOT NULL,
		cloud_cover real NOT NULL,
		wrs_path smallint NOT NULL,
		wrs_row smallint NOT NULL,
		scene_url text COLLATE pg_catalog."default" NOT NULL,
		bounds geometry NOT NULL,
		CONSTRAINT "scenes_pk_productId" PRIMARY KEY (product_id)
	)
	WITH (
		OIDS = FALSE
	);
		`)

	return err
}

func populateWRSPaths(tx *sql.Tx) error {
	//Create the insert statement.
	insertStatement, err := tx.Prepare(
		`
		INSERT INTO wrs2paths
		(path, row, center, boundary)
		VALUES
		(
			$1,
			$2,
			ST_SetSRID(ST_MakePoint($3, $4),4326),
			ls_clampedToWorldBoundary(
			ST_SetSRID(
				ST_MakePolygon(
					ST_MakeLine(
						ARRAY[
							ST_MakePoint($5, $6),
							ST_MakePoint($7, $8),
							ST_MakePoint($9, $10),
							ST_MakePoint($11, $12),
							ST_MakePoint($5, $6)
							]
						)
					), 4326)
					))
		`)
	if err != nil {
		return err
	}

	strReader := strings.NewReader(wrsCornerPointsCSV)
	reader := csv.NewReader(strReader)

	reader.ReuseRecord = true

	//Discard the line with the column names.
	line, err := reader.Read()

	numberRead := 0
	for err == nil {
		line, err = reader.Read()
		numberRead++
		if err == nil {
			_, err = insertStatement.Exec(
				line[0],          //path
				line[1],          //row
				line[3], line[2], //center lon, lat
				line[5], line[4], //upper left
				line[7], line[6], //upper right
				line[11], line[10], //lower right
				line[9], line[8]) //lower left
		}
	}

	//For debug.
	log.Println("Number read:", numberRead)
	log.Println(line)

	if err == io.EOF {
		//Read to the end of the file. Success!
		err = nil
	}

	return err
}
