package landsataws

import (
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"time"

	csvcolumnmap "github.com/venicegeo/bf-ia-broker/landsatAws/csvColumnMap"
)

type insertStats struct {
	NumberSuccess int
	NumberError   int
}

//csvValueConverter is used to transform the values from the csv file into
//the parameter values that will be injected into the SQL INSERT statement
type csvValueConverter func(map[string]string) (interface{}, error)

//Ingest reads from the stream as a CSV and inserts/updates database records for scenes.
func Ingest(reader io.Reader, database *sql.DB) {
	csvReader := csv.NewReader(reader)
	firstRow, err := csvReader.Read() //discard the first row.
	if err != nil {
		log.Fatal("Error reading first line.")
	}

	colMap, err := csvcolumnmap.New(columnNames, firstRow)
	if err != nil {
		log.Fatal("Error extracting column names.")
	}

	ingest(csvReader, insertSceneStatement, colMap, columnConverters, database)
}

//ingest reads the csv file and populates that database
func ingest(
	sceneCsv *csv.Reader,
	insertStatement string,
	columnMap csvcolumnmap.CsvColumnMap,
	converters []csvValueConverter,
	db *sql.DB) {

	//Create the prepared statement that will be used to insert records.
	stmt, err := db.Prepare(insertStatement)
	if err != nil {
		log.Fatal("Prepare statement failed.")
	}
	defer stmt.Close()

	//Create the map that allows values to be found by column name
	valueMap := columnMap.CreateValueMap()

	var rawLineValues []string
	var csvErr error
	sceneCsv.ReuseRecord = true

	var stats insertStats
	startTime := time.Now()
	lastProgressLogTime := time.Now()
	progressLogInterval := time.Duration(time.Second * 30)
CSVLoop:
	for {
		if time.Since(lastProgressLogTime) > progressLogInterval {
			log.Printf("Ingest Progress: %+v", stats)
			lastProgressLogTime = time.Now()
		}

		rawLineValues, csvErr = sceneCsv.Read()
		switch csvErr {
		case nil:
			columnMap.UpdateMap(rawLineValues, valueMap)
			err = executeInsert(stmt, valueMap, converters)
			if err != nil {
				stats.NumberError++
				log.Println("Error inserting scene into db.", err, rawLineValues)
			} else {
				stats.NumberSuccess++
			}
		case io.EOF:
			break CSVLoop
		default:
			log.Println("Error reading csv line:", csvErr, rawLineValues)
		}
	}
	log.Printf("Ingest Complete: %+v", stats)
	log.Printf("Ingest took %s", time.Since(startTime))

	doDatabaseMaintenance(db)
}

//doDatabaseMaintenance performs any maintenance that should be done
//after the import operation, e.g. rebuilding indexes
func doDatabaseMaintenance(database *sql.DB) {
	log.Println("Starting database maintenance.")
	_, err := database.Exec(databaseMaintenanceStatement)
	if err != nil {
		log.Println("Error during database maintenance.", err)
	}
	log.Println("Database maintenance complete.")
}

//excecuteInsert submits the insert statement to the database driver.
func executeInsert(
	statement *sql.Stmt,
	valueMap map[string]string,
	converters []csvValueConverter) error {
	var err error
	dbValues := make([]interface{}, len(converters))

	for idx, conv := range converters {
		dbValues[idx], err = conv(valueMap)

		if err != nil {
			log.Printf("Failed to convert field %v from values %v.", idx, valueMap)
			break
		}
	}

	if err == nil {
		_, err = statement.Exec(dbValues...)
	}

	return err
}
