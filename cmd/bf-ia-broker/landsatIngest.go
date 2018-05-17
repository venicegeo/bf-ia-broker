package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/venicegeo/bf-ia-broker/landsataws"

	_ "github.com/lib/pq"
	cli "gopkg.in/urfave/cli.v1"
)

const connectionStringEnv = "postgresConnectionString"
const scenesFileEnv = "scenesCsvUrl"
const ingestFrequencyEnv = "ingest_frequency"
const defaultIngestFrequency = 24 * time.Hour

//landsatIngestAction starts the worker process and an http server
func landsatIngestAction(*cli.Context) {
	portStr := getPortStr()

	scenesURL := os.Getenv(scenesFileEnv)
	scenesIsGzip := strings.HasSuffix(strings.ToLower(scenesURL), "gz")
	importer := landsataws.NewImporter(scenesURL, scenesIsGzip, getDbConnection)

	//Create the channel that sends the star/stop messages to the Importer.
	messageChan := make(chan string, 5) //small buffer.

	//Start the sleep/ingest loop.
	go importer.ImportWhile(messageChan, getTimerDuration())

	//Set up an http router
	router := mux.NewRouter()
	router.HandleFunc("/ingest/", func(resp http.ResponseWriter, req *http.Request) {
		handleImportStatus(importer, resp, req)
	})
	router.HandleFunc("/ingest/start", func(resp http.ResponseWriter, req *http.Request) {
		handleForceStartIngest(importer, messageChan, resp, req)
	})
	router.HandleFunc("/ingest/cancel", func(resp http.ResponseWriter, req *http.Request) {
		handleCancel(importer, messageChan, resp, req)
	})

	log.Println("Listening on port", portStr)
	log.Fatal(http.ListenAndServe(portStr, router))
}

//handleImportStatus requests the status from the importer and writes it out.
func handleImportStatus(imp *landsataws.Importer, writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(writer, imp.GetStatus())
}

//handleForceStartIngest sends a "begin" message to the importer and returns the new status to the user.
func handleForceStartIngest(imp *landsataws.Importer, messageChan chan<- string, writer http.ResponseWriter, req *http.Request) {
	select {
	case messageChan <- landsataws.BeginIngestJobMessage:
		fmt.Fprintln(writer, "Begin job request submitted.")
	default:
		fmt.Fprintln(writer, "Error submitting request.")
	}
	fmt.Fprintln(writer, imp.GetStatus())
}

//handleCancel sends a "cancel" message to the importer and returns the new status to the user.
func handleCancel(imp *landsataws.Importer, cancelChan chan<- string, writer http.ResponseWriter, req *http.Request) {
	select {
	case cancelChan <- landsataws.AbortIngestJobMessage:
		fmt.Fprintln(writer, "Cancel request submitted.")
	default:
		fmt.Fprintln(writer, "Error submitting cancel request.")
	}
	fmt.Fprintln(writer, imp.GetStatus())
}

//getDbConnection opens a new database connection.
func getDbConnection() (*sql.DB, error) {
	connStr := os.Getenv(connectionStringEnv)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

func getTimerDuration() time.Duration {
	duration, _ := time.ParseDuration(os.Getenv(ingestFrequencyEnv))

	if duration < (time.Minute) {
		log.Printf("Specified duration of %v is too small. Setting to default.", duration)
		duration = defaultIngestFrequency
	}

	return duration
}
