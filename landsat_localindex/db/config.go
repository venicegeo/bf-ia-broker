package db

const productIDColumn string = "productId"
const captureDateColumn string = "acquisitionDate"
const cloudCoverColumn string = "cloudCover"
const wrsPathColumn string = "path"
const wrsRowColumn string = "row"
const downloadURLColumn = "download_url"

const insertSceneStatement = `
INSERT INTO scenes as s (
	product_id,
	acquisition_date,
	cloud_cover,
	wrs_path,
	wrs_row,
	scene_url,
	bounds)
VALUES
(
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	(SELECT boundary FROM wrs2paths WHERE
	path=$4 AND row=$5 LIMIT 1)
)
	ON CONFLICT (product_id) DO UPDATE
	SET scene_url =	$6
	WHERE s.scene_url <> $6
	`

const databaseMaintenanceStatement = `
	VACUUM ANALYZE scenes
`

//columnNames should contain an entry for any column used in a columnCoverter.
var columnNames = []string{
	productIDColumn,
	captureDateColumn,
	cloudCoverColumn,
	wrsPathColumn,
	wrsRowColumn,
	downloadURLColumn}

//columnConverters transform the raw values from the csv file into the values of the
//parameters used in the insert SQL statement.
//NOTE: Since the database can do most necessary parsing this may all be trivial.
var columnConverters = []csvValueConverter{
	func(vals map[string]string) (interface{}, error) { return vals[productIDColumn], nil },
	func(vals map[string]string) (interface{}, error) { return vals[captureDateColumn], nil },
	func(vals map[string]string) (interface{}, error) { return vals[cloudCoverColumn], nil },
	func(vals map[string]string) (interface{}, error) { return vals[wrsPathColumn], nil },
	func(vals map[string]string) (interface{}, error) { return vals[wrsRowColumn], nil },
	func(vals map[string]string) (interface{}, error) { return vals[downloadURLColumn], nil }}
