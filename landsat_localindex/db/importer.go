package db

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/venicegeo/bf-ia-broker/util"
)

//ConnectionProvider is a function that can provide a database connection.
type ConnectionProvider func(util.LogContext) (*sql.DB, error)

//Importer manages the state for an ingest job.
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
//To close quickly, send landsataws.AbortIngestJobMessage on messageChan before closing it.
func (imp *Importer) ImportWhile(messageChan <-chan string, maxTimeBetweenJobs time.Duration) {
	log.Println("Job loop started with frequency", maxTimeBetweenJobs)

	previousStatus := "\tNone"

	scheduleTimer := time.NewTimer(maxTimeBetweenJobs)
	nextScheduledStartTime := time.Now().Add(maxTimeBetweenJobs)

	var startJob bool
	for {
		startJob = false

		//Wait for a start message.
		//Also, status is reported cooperatively. (maybe not super elegant) so deal with any requests while we wait.
		select {
		case <-scheduleTimer.C:
			log.Println("Maximum time between jobs elapsed.")
			startJob = true
		case msg, ok := <-messageChan:
			if !ok {
				return //The message channel has been closed. Exit.
			}
			switch msg {
			case BeginIngestJobMessage:
				//The user has sent a start message. Start a job.
				log.Println("User requested job start.")
				startJob = true
			default:
				//ignore this message. We only want ones for "begin".
			}
		case respChan := <-imp.statusChan:
			//The user has sent a request for the current status.
			select {
			//Try to send a response on the provided channel.
			case respChan <- fmt.Sprintf("%v\nStatus: Sleeping until %v\nPrevious job:\n%v",
				time.Now().Format("Mon Jan _2 15:04:05 2006"),
				nextScheduledStartTime.Format("Mon Jan _2 15:04:05 2006"),
				previousStatus): //good
			default:
				//Could not send immediately. We'll ignore it.
			}
		}

		if startJob {
			log.Println("Starting job.")
			//Do the actual import.
			previousStatus = imp.Import(messageChan)

			//Reset the timer.
			scheduleTimer.Stop()
			//Rather than keep track of whether we've recieved on the timer channel (maybe that's how we got here),
			//we'll just drain it in a general way.
		TimerDrainLoop:
			for {
				select {
				case <-scheduleTimer.C: //good, discard
				default:
					//Channel is empty. We're done.
					break TimerDrainLoop
				}
			}

			//This simple implementation just sets the timer to some duration in the future.
			//It might be desireble to add a reference time, e.g. so the timer only triggers
			//during low-usage hours.
			scheduleTimer.Reset(maxTimeBetweenJobs)
			nextScheduledStartTime = time.Now().Add(maxTimeBetweenJobs)
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

	//If the user has indicated that the csv uses gzip, then
	//wrap the the reader.
	if imp.scenesIsGzip {
		archiveReader, zipErr := gzip.NewReader(mainReader)
		if zipErr != nil {
			log.Fatal("Error opening gzip archive.", zipErr)
		}
		defer archiveReader.Close()
		mainReader = archiveReader
	}

	//Database connection is opened right before the ingest, and closed
	//immediately after.
	database, err := imp.dbConnProvider(&util.BasicLogContext{})
	if err != nil {
		log.Fatal("Could not open database connection.")
	}
	defer database.Close()

	return imp.Ingest(mainReader, database, messageChan)
}

func openReader(scenesURL string) (io.ReadCloser, error) {
	//If this looks like a url then try to download it.
	if strings.HasPrefix(scenesURL, "http://") || strings.HasPrefix(scenesURL, "https://") {
		log.Println("Requesting url:", scenesURL)
		archiveResponse, netErr := http.Get(scenesURL)
		if netErr != nil {
			return nil, netErr
		}
		defer archiveResponse.Body.Close()

		//Download the whole body so we don't need to keep the connection open
		bodyData, _ := ioutil.ReadAll(archiveResponse.Body)

		return ioutil.NopCloser(bytes.NewBuffer(bodyData)), nil
	}

	//Treat this as a file.
	cleanPath := filepath.Clean(scenesURL)
	log.Println("Opening file", cleanPath)
	file, err := os.Open(cleanPath)
	return file, err
}
