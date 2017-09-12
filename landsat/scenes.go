package landsat

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/venicegeo/bf-ia-broker/util"
)

type sceneMapRecord struct {
	awsFolderURL string
	filePrefix   string
}


var sceneMap = map[string]sceneMapRecord{}

// SceneMapIsReady contains a flag of whether the scene map has been loaded yet
var SceneMapIsReady = false

// UpdateSceneMap updates the global scene map from a remote source
func UpdateSceneMap(ctx util.LogContext) (err error) {
	landSatHost := util.GetLandsatHost()
	sceneListURL := fmt.Sprintf("%s/c1/L8/scene_list.gz", landSatHost)

	util.LogAudit(ctx, util.LogAuditInput{Actor: "anon user", Action: "GET", Actee: sceneListURL, Message: "Importing scene list", Severity: util.INFO})
	c := util.HTTPClient()
	response, err := c.Get(sceneListURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		err = fmt.Errorf("Non-200 response code: %d", response.StatusCode)
		return
	}

	rawReader := response.Body
	gzipReader, err := gzip.NewReader(rawReader)
	if err != nil {
		return
	}

	csvReader := csv.NewReader(gzipReader)
	newSceneMap := map[string]sceneMapRecord{}
doneReading:
	for {
		record, readErr := csvReader.Read()
		switch readErr {
		case nil:
			// First column contains file prefix
			filePrefix := record[0]
			// Second column contains scene ID
			id := record[1]
			// Last column contains URL
			url := record[len(record)-1]
			// Strip the "index.html" file name to just get the directory path
			lastSlash := strings.LastIndex(url, "/")
			url = url[:lastSlash+1]

			newSceneMap[id] = sceneMapRecord{filePrefix: filePrefix, awsFolderURL: url}
		case io.EOF:
			break doneReading
		default:
			err = readErr
			return
		}
	}

	sceneMap = newSceneMap
	SceneMapIsReady = true
	util.LogAudit(ctx, util.LogAuditInput{Actor: "anon user", Action: "GET", Actee: sceneListURL, Message: "Imported scene list", Severity: util.INFO})
	return nil
}

// UpdateSceneMapAsync runs UpdateSceneMap asynchronously, returning
// completion signals via channels
func UpdateSceneMapAsync(ctx util.LogContext) (done chan bool, errored chan error) {
	done = make(chan bool)
	errored = make(chan error)
	go func() {
		err := UpdateSceneMap(ctx)
		if err == nil {
			done <- true
		} else {
			errored <- err
		}
		close(done)
		close(errored)
	}()
	return
}

// UpdateSceneMapOnTicker updates the scene map on a loop with a delay of
// a given duration. It logs any errors using the given LogContext
func UpdateSceneMapOnTicker(d time.Duration, ctx util.LogContext) {
	ticker := time.NewTicker(d)
	for {
		done, errored := UpdateSceneMapAsync(ctx)
		select {
		case <-done:
		case err := <-errored:
			util.LogAlert(ctx, "Failed to update scene ID to URL map: "+err.Error())
		}
		<-ticker.C
	}
}

// GetSceneFolderURL returns the AWS S3 URL at which the scene files for this
// particular scene are available
func GetSceneFolderURL(sceneID string, dataType string) (folderURL string, filePrefix string, err error) {
	if !IsValidLandSatID(sceneID) {
		return "", "", fmt.Errorf("Invalid scene ID: %s", sceneID)
	}

	isPreC1 := IsPreCollectionDataType(dataType)
	isC1 := IsCollection1DataType(dataType)
	if !(isPreC1 || isC1) {
		return "", "", errors.New("Unknown LandSat data type: " + dataType)
	}

	if isC1 {
		if !SceneMapIsReady {
			return "", "", errors.New("Scene map is not ready yet")
		}
		if record, ok := sceneMap[sceneID]; ok {
			return record.awsFolderURL, record.filePrefix, nil
		}
	}

	if isPreC1 {
		return formatPreCollectionIDToURL(sceneID), sceneID, nil
	}

	return "", "", fmt.Errorf("Scene not found with ID: %s, dataType: %s", sceneID, dataType)
}