package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	db "github.com/venicegeo/bf-ia-broker/landsat_localindex/db"

	_ "github.com/lib/pq"
	cli "gopkg.in/urfave/cli.v1"
)

const scenesFileEnv = "LANDSAT_INDEX_SCENES_URL"
const ingestFrequencyEnv = "LANDSAT_INGEST_FREQUENCY"
const defaultIngestFrequency = 24 * time.Hour

//calls the ingest worker a single time without scheduling
func landsatIngestOnceAction(*cli.Context) {
	//Start the sleep/ingest loop.
	go importer.Import(nil)
}

//landsatIngestAction starts the worker process and an http server
func landsatIngestScheduleAction(*cli.Context) {
	portStr := getPortStr()

	scenesURL := os.Getenv(scenesFileEnv)
	scenesIsGzip := strings.HasSuffix(strings.ToLower(scenesURL), "gz")
	importer := db.NewImporter(scenesURL, scenesIsGzip, getDbConnectionFunc)

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
func handleImportStatus(imp *db.Importer, writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(writer, imp.GetStatus())
}

//handleForceStartIngest sends a "begin" message to the importer and returns the new status to the user.
func handleForceStartIngest(imp *db.Importer, messageChan chan<- string, writer http.ResponseWriter, req *http.Request) {
	select {
	case messageChan <- db.BeginIngestJobMessage:
		fmt.Fprintln(writer, "Begin job request submitted.")
	default:
		fmt.Fprintln(writer, "Error submitting request.")
	}
	fmt.Fprintln(writer, imp.GetStatus())
}

//handleCancel sends a "cancel" message to the importer and returns the new status to the user.
func handleCancel(imp *db.Importer, cancelChan chan<- string, writer http.ResponseWriter, req *http.Request) {
	select {
	case cancelChan <- db.AbortIngestJobMessage:
		fmt.Fprintln(writer, "Cancel request submitted.")
	default:
		fmt.Fprintln(writer, "Error submitting cancel request.")
	}
	fmt.Fprintln(writer, imp.GetStatus())
}

func getTimerDuration() time.Duration {
	duration, _ := time.ParseDuration(os.Getenv(ingestFrequencyEnv))

	if duration < (time.Minute) {
		log.Printf("Specified duration of %v is too small. Setting to default.", duration)
		duration = defaultIngestFrequency
	}

	return duration
}
