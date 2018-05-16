package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/venicegeo/bf-ia-broker/landsataws"

	_ "github.com/lib/pq"
	cli "gopkg.in/urfave/cli.v1"
)

const connectionStringEnv = "postgresConnectionString"
const scenesFileEnv = "scenesCsvUrl"

//landsatIngestAction starts the worker process and an http server
func landsatIngestAction(*cli.Context) {
	portStr := getPortStr()

	scenesURL := os.Getenv(scenesFileEnv)
	scenesIsGzip := strings.HasSuffix(strings.ToLower(scenesURL), "gz")
	importer := landsataws.NewImporter(scenesURL, scenesIsGzip, getDbConnection)

	messageChan := make(chan string, 5) //small buffer. 1 is probably sufficient.
	//scheduleTicker := time.NewTicker()  //time.NewTicker(24 * time.Hour)

	//Start the sleep/ingest loop.
	go importer.ImportWhile(messageChan, 2*time.Minute)
	//go sendStartMessageOnTick(scheduleTicker.C, messageChan)

	router := landsataws.NewRouter()
	router.HandleFunc("/ingest/", func(resp http.ResponseWriter, req *http.Request) { handleImportStatus(importer, resp, req) })
	router.HandleFunc("/ingest/forceStart", func(resp http.ResponseWriter, req *http.Request) {
		handleForceStartIngest(importer, messageChan, resp, req)
	})
	router.HandleFunc("/ingest/cancelJob", func(resp http.ResponseWriter, req *http.Request) { handleCancel(importer, messageChan, resp, req) })

	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe(portStr, router))

}

// func sendStartMessageOnTick(producer <-chan time.Time, consumer chan<- string) {
// 	for range producer {
// 		consumer <- landsataws.BeginIngestJobMessage
// 	}
// }

func handleImportStatus(imp *landsataws.Importer, writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(writer, imp.GetStatus())
}

func handleForceStartIngest(imp *landsataws.Importer, messageChan chan<- string, writer http.ResponseWriter, req *http.Request) {
	select {
	case messageChan <- landsataws.BeginIngestJobMessage:
		fmt.Fprintln(writer, "Begin job request submitted.")
	default:
		fmt.Fprintln(writer, "Error submitting request.")
	}
	fmt.Fprintln(writer, imp.GetStatus())
}

func handleCancel(imp *landsataws.Importer, cancelChan chan<- string, writer http.ResponseWriter, req *http.Request) {
	select {
	case cancelChan <- landsataws.AbortIngestJobMessage:
		fmt.Fprintln(writer, "Cancel request submitted.")
	default:
		fmt.Fprintln(writer, "Error submitting cancel request.")
	}
	fmt.Fprintln(writer, imp.GetStatus())
}

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
