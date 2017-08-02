package main

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/bf-ia-broker/landsat"
)

const badLandSatID = "X_NOT_LANDSAT_X"
const oldLandSatID = "LC8123456890"
const newLandSatID = "LC08_L1TP_012029_20170213_20170415_01_T1"
const newLandSatURL = "https://s3-us-west-2.fakeamazonaws.dummy/thisiscorrect/"
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

func TestServe_CallsLaunchServer(t *testing.T) {
	success := make(chan bool)
	launchServer = func(portStr string, router *mux.Router) { // Mock
		success <- true
	}
	timer := time.NewTimer(1 * time.Second)

	go serve()

	select {
	case <-success:
	case <-timer.C:
		assert.Fail(t, "launchServer not called within 1 second of serve()")
	}
}

func TestServe_SeedsLandSatC1Mappings(t *testing.T) {
	launchServer = func(portStr string, router *mux.Router) {} // Mock

	go serve()
	<-time.NewTimer(1 * time.Second).C

	assert.True(t, landsat.SceneMapIsReady, "LandSat scene map took more than 1 second to load")
}

func TestServe_BaseHealthCheckEndpoint(t *testing.T) {
	success := make(chan bool)
	launchServer = func(portStr string, router *mux.Router) { // Mock
		req := httptest.NewRequest("GET", "/", strings.NewReader(""))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		responseBody, _ := ioutil.ReadAll(response.Result().Body)
		success <- (string(responseBody) == "Hi")
	}

	timer := time.NewTimer(1 * time.Second)

	go serve()

	select {
	case <-success:
	case <-timer.C:
		assert.Fail(t, "launchServer not called within 1 second of serve()")
	}
}
