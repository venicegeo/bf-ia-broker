package landsataws

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//ConnectionProvider is a function that can provide a database connection.
type ConnectionProvider func() (*sql.DB, error)

//Importer manages the state for an import job.
//Mainly useful when launching the job on an interval.
type Importer struct {
	scenesURL      string
	scenesIsGzip   bool
	dbConnProvider ConnectionProvider
	statusChan     chan chan string
}

//NewImporter intializes a new importer.
func NewImporter(
	url string,
	useGzip bool,
	dbConnProvider ConnectionProvider) *Importer {
	return &Importer{
		scenesURL:      url,
		scenesIsGzip:   useGzip,
		dbConnProvider: dbConnProvider,
		statusChan:     make(chan chan string, 10)}
}

//ImportWhile peforms the Ingest() task and waits for a channel.
//Note: this is blocking
//The function will exit when messageChan is closed and any in-progress jobs complete.
//To close quickly, send a stop message on messageChan.
func (imp *Importer) ImportWhile(messageChan <-chan string, maxTimeBetweenJobs time.Duration) {
	previousStatus := "\tNone"
	var nextScheduledStartTime time.Time
	var scheduleTimer *time.Timer

	for {
		if scheduleTimer == nil {
			scheduleTimer = time.NewTimer(maxTimeBetweenJobs)
			nextScheduledStartTime = time.Now().Add(maxTimeBetweenJobs)
		}

		select {
		case <-scheduleTimer.C:
			scheduleTimer = nil
			previousStatus = imp.Import(messageChan)
		case msg, ok := <-messageChan:
			if !ok {
				return //The message channel has been closed.
			}
			switch msg {
			case BeginIngestJobMessage:
				scheduleTimer = nil
				previousStatus = imp.Import(messageChan)
			default:
				//ignore this message. We only want ones for begin.
			}
		case respChan := <-imp.statusChan:
			select {
			case respChan <- fmt.Sprintf("%v\nStatus: Sleeping until %v\nPrevious job:\n%v",
				time.Now().Format("Mon Jan _2 15:04:05 2006"),
				nextScheduledStartTime.Format("Mon Jan _2 15:04:05 2006"),
				previousStatus): //good
			default: //ignore
			}
		}
	}
}

//GetStatus is a thread safe way to get information about the import operation.
func (imp *Importer) GetStatus() string {
	responseChan := make(chan string, 1) //Must have a buffer. reportStatus won't wait if it can't send.
	imp.statusChan <- responseChan
	status := <-responseChan
	return status
}

//Import performs the actual read and update.
func (imp *Importer) Import(messageChan <-chan string) (result string) {
	var mainReader io.Reader
	sourceReader, err := openReader(imp.scenesURL)
	if err != nil {
		log.Fatal("Could not open the source file/url.")
	}
	defer sourceReader.Close()
	mainReader = sourceReader

	if imp.scenesIsGzip {
		archiveReader, zipErr := gzip.NewReader(mainReader)
		if zipErr != nil {
			log.Fatal("Error opening gzip archive.", zipErr)
		}
		defer archiveReader.Close()
		mainReader = archiveReader
	}

	database, err := imp.dbConnProvider()
	if err != nil {
		log.Fatal("Could not open database connection.")
	}
	defer database.Close()

	return imp.Ingest(mainReader, database, messageChan)
}

func openReader(scenesURL string) (io.ReadCloser, error) {
	if strings.HasPrefix(scenesURL, "http://") || strings.HasPrefix(scenesURL, "https://") {
		log.Println("Requesting url:", scenesURL)
		archiveResponse, netErr := http.Get(scenesURL)
		if netErr != nil {
			return nil, netErr
		}
		return archiveResponse.Body, nil
	}

	//Treat this as a file.
	cleanPath := filepath.Clean(scenesURL)
	log.Println("Opening file", cleanPath)
	file, err := os.Open(cleanPath)
	return file, err
}
