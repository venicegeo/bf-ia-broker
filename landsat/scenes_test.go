package landsat

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const badLandSatID = "X_NOT_LANDSAT_X"
const oldLandSatID = "LC8123456890"
const newLandSatID = "LC08_L1TP_012029_20170213_20170415_01_T1"
const newLandSatURL = "https://s3-us-west-2.fakeamazonaws.dummy/thisiscorrect/index.html"
const missingNewLandSatID = "LC08_L1TP_012029_20180213_20170415_01_T1"

var sampleSceneMapCSV = []byte(newLandSatID +
	",LC81490392017101LGN00,2017-04-11 05:36:29.349932,0.0,L1TP,149,39,29.22165,72.41205,31.34742,74.84666," +
	newLandSatURL)

type mockAWSHandler struct{}

func (h mockAWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gzipWriter := gzip.NewWriter(w)
	gzipWriter.Write(sampleSceneMapCSV)
	gzipWriter.Close()
}

func TestMain(m *testing.M) {
	mockAWSServer := httptest.NewServer(mockAWSHandler{})
	defer mockAWSServer.Close()
	os.Setenv("LANDSAT_HOST", mockAWSServer.URL)
	code := m.Run()
	os.Exit(code)
}

func TestGetSceneFolderURL_BadIDs(t *testing.T) {
	_, err := GetSceneFolderURL(badLandSatID)
	assert.NotNil(t, err, "Invalid LandSat ID did not cause an error")
	assert.Contains(t, err.Error(), "Invalid scene ID")

	_, err = GetSceneFolderURL(missingNewLandSatID)
	assert.NotNil(t, err, "Scene map not ready did not cause an error")
	assert.Contains(t, err.Error(), "not ready")

	UpdateSceneMap(mockLogContext{})
	_, err = GetSceneFolderURL(missingNewLandSatID)
	assert.NotNil(t, err, "Missing scene ID did not cause an error")
	assert.Contains(t, err.Error(), "not found")
}

func TestGetSceneFolderURL_OldSceneID(t *testing.T) {
	url, err := GetSceneFolderURL(oldLandSatID)
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, url, fmt.Sprintf(oldLandSatAWSURL, "123", "456", oldLandSatID, ""))
}

func TestGetSceneFolderURL_NewSceneID(t *testing.T) {
	UpdateSceneMap(mockLogContext{})
	url, err := GetSceneFolderURL(newLandSatID)
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, "https://s3-us-west-2.fakeamazonaws.dummy/thisiscorrect/", url)
}

func TestUpdateSceneMapAsync_Success(t *testing.T) {
	done, errored := UpdateSceneMapAsync(mockLogContext{})
	expireTimer := time.NewTimer(1 * time.Second)
	select {
	case <-done:
		return
	case err := <-errored:
		assert.Fail(t, err.Error())
	case <-expireTimer.C:
		assert.Fail(t, "Timed out")
	}
}

type mockLogContext struct{}

func (ctx mockLogContext) AppName() string    { return "bf-ia-broker TESTING" }
func (ctx mockLogContext) SessionID() string  { return "test-session" }
func (ctx mockLogContext) LogRootDir() string { return "/tmp" }
