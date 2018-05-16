package landsataws

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	csvcolumnmap "github.com/venicegeo/bf-ia-broker/landsatAws/csvColumnMap"
)

//BeginIngestJobMessage is sent on a channel to start an ingest job.
const BeginIngestJobMessage = "start"

//AbortIngestJobMessage is sent on a channel to stop an in-progress job.
const AbortIngestJobMessage = "stop"

type jobStats struct {
	NumberAddedOrUpdated int
	NumberSkipped        int
	NumberError          int
	StartTime            time.Time
	EndTime              time.Time
	CanceledByUser       bool
}

func (stats *jobStats) String() string {
	return fmt.Sprintf(`
		Start:	%v
		End:	%v
		Canceled: %v
		#Added:		%v
		#Skipped:	%v
		#Error:		%v
		`,
		stats.StartTime.Format("Mon Jan _2 15:04:05 2006"),
		stats.EndTime.Format("Mon Jan _2 15:04:05 2006"),
		stats.CanceledByUser,
		stats.NumberAddedOrUpdated,
		stats.NumberSkipped,
		stats.NumberError)
}

//csvValueConverter is used to transform the values from the csv file into
//the parameter values that will be injected into the SQL INSERT statement
type csvValueConverter func(map[string]string) (interface{}, error)

//Ingest reads from the stream as a CSV and inserts/updates database records for scenes.
func (imp *Importer) Ingest(reader io.Reader, database *sql.DB, cancelChan <-chan string) (result string) {
	csvReader := csv.NewReader(reader)
	firstRow, err := csvReader.Read() //discard the first row.
	if err != nil {
		log.Fatal("Error reading first line.")
	}

	colMap, err := csvcolumnmap.New(columnNames, firstRow)
	if err != nil {
		log.Fatal("Error extracting column names.")
	}

	return imp.ingest(csvReader, insertSceneStatement, colMap, columnConverters, database, cancelChan)
}

//ingest reads the csv file and populates that database
func (imp *Importer) ingest(
	sceneCsv *csv.Reader,
	insertStatement string,
	columnMap csvcolumnmap.CsvColumnMap,
	converters []csvValueConverter,
	db *sql.DB,
	cancelChan <-chan string) (result string) {

	//Create the prepared statement that will be used to insert records.
	stmt, err := db.Prepare(insertStatement)
	if err != nil {
		log.Fatal("Prepare statement failed.", err)
	}
	defer stmt.Close()

	//Create the map that allows values to be found by column name
	valueMap := columnMap.CreateValueMap()

	var rawLineValues []string
	var csvErr error
	sceneCsv.ReuseRecord = true

	var stats jobStats
	stats.StartTime = time.Now()
	lastProgressLogTime := time.Now()
	progressLogInterval := time.Duration(time.Second * 30)

CSVLoop:
	for {
		//Check whether the user has requested cancelation.
		if abort := drainMessages(cancelChan); abort {
			log.Println("Ingest job canceld by user.")
			stats.CanceledByUser = true
			break CSVLoop
		}

		//Report the status to anyone waiting for it.
		drainStatusChannel(imp.statusChan, &stats)

		//Occasionally emit progess to the log stream
		if time.Since(lastProgressLogTime) > progressLogInterval {
			log.Printf("Ingest Progress: Added:%v Skipped:%v Error:%v", stats.NumberAddedOrUpdated, stats.NumberSkipped, stats.NumberError)
			lastProgressLogTime = time.Now()
		}

		//Read a line from the CSV file.
		rawLineValues, csvErr = sceneCsv.Read()
		switch csvErr {
		case nil:
			//Parse the values.
			columnMap.UpdateMap(rawLineValues, valueMap)
			//Insert the values into the database.
			rowsAffected, err := executeInsert(stmt, valueMap, converters)
			if err != nil {
				stats.NumberError++
				log.Println("Error inserting scene into db.", err, rawLineValues)
			} else {
				//No error during database insert.
				stats.NumberAddedOrUpdated += rowsAffected
				stats.NumberSkipped += (1 - rowsAffected)
			}
		case io.EOF:
			//Read to the end of the file. Exit the loop.
			break CSVLoop
		default:
			//Something went wrong reading the line from the file. Possibly formatting.
			//Just log this and move along.
			log.Println("Error reading csv line:", csvErr, rawLineValues)
			stats.NumberError++
		}
	}

	//Clear the status requests before doing the long-running operation.
	drainStatusChannel(imp.statusChan, &stats)
	doDatabaseMaintenance(db)

	stats.EndTime = time.Now()
	log.Printf("Ingest Complete: %v", stats.String())
	log.Printf("Ingest took %s", stats.EndTime.Sub(stats.StartTime))

	return fmt.Sprintf("%v", stats.String())
}

//draingMessages reads all the messages from the channel looking for
//any abort messages.
//All other messages will be ignored and discarded.
func drainMessages(messageChan <-chan string) (abortRequested bool) {
	abortRequested = false
	for {
		select {
		case msg := <-messageChan:
			abortRequested = abortRequested || (msg == AbortIngestJobMessage)
		default:
			return
		}
	}
}

//reportStatus drains the status request channel
//and sends back a status string
func drainStatusChannel(statusChan <-chan chan string, stats *jobStats) {
	for {
		select {
		case resp := <-statusChan:
			if resp != nil {
				select {
				case resp <- fmt.Sprintf("%v\nIn progress\n%v", time.Now().Format("Mon Jan _2 15:04:05 2006"), stats.String()): //good
				default: //can't send. ignore this request.
				}
			}
		default:
			return
		}
	}
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
	converters []csvValueConverter) (int, error) {

	var err error
	dbValues := make([]interface{}, len(converters))

	for idx, conv := range converters {
		dbValues[idx], err = conv(valueMap)

		if err != nil {
			log.Printf("Failed to convert field %v from values %v.", idx, valueMap)
			break
		}
	}

	if err != nil {
		return 0, err
	}

	var rowsAffected int64 // zero
	result, err := statement.Exec(dbValues...)
	if err != nil {
		rowsAffected, err = result.RowsAffected()
	}

	return int(rowsAffected), err
}
