// Copyright 2018, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/venicegeo/bf-ia-broker/util"
)

const (
	badLandSatID               = "X_NOT_LANDSAT_X"
	goodLandSatID1             = "LC81234567890"
	goodLandSatID2             = "LC89876543210"
	goodLandSatIDNotInSceneMap = "LC81029384756"
	missingLandSatID           = "LC8123456000"
	collection1ID1             = "LONG_COLLECTION_1_ID_1"
	collection1ID2             = "LONG_COLLECTION_1_ID_2"
	l1tpLandSatURL             = "https://s3-us-west-2.fakeamazonaws.dummy/thisiscorrect/index.html"
	l1gtLandSatURL             = "https://s3-us-west-2.fakeamazonaws.dummy/thisisalsocorrect/index.html"
	l1tDataType                = "L1T"
	l1gtDataType               = "L1GT"
	l1tpDataType               = "L1TP"
	badDataType                = "BOGUS"
)

var sampleSceneMapCSV = []byte(
	collection1ID1 + "," + goodLandSatID1 + ",2017-04-11 05:36:29.349932,0.0," + l1tpDataType + ",149,39,29.22165,72.41205,31.34742,74.84666," + l1tpLandSatURL + "\n" +
		collection1ID2 + "," + goodLandSatID2 + ",2017-04-11 05:36:29.349932,0.0," + l1gtDataType + ",149,39,29.22165,72.41205,31.34742,74.84666," + l1gtLandSatURL + "\n",
)

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
	_, _, err := GetSceneFolderURL(badLandSatID, l1tpDataType)
	assert.NotNil(t, err, "Invalid LandSat ID did not cause an error")
	assert.Contains(t, err.Error(), "Invalid scene ID")

	_, _, err = GetSceneFolderURL(goodLandSatID1, l1tpDataType)
	assert.NotNil(t, err, "Scene map not ready did not cause an error")
	assert.Contains(t, err.Error(), "not ready")

	UpdateSceneMap(mockLogContext{})
	_, _, err = GetSceneFolderURL(missingLandSatID, l1tpDataType)
	assert.NotNil(t, err, "Missing scene ID did not cause an error")
	assert.Contains(t, err.Error(), "not found")
}

func TestGetSceneFolderURL_BadDataType(t *testing.T) {
	UpdateSceneMap(mockLogContext{})
	_, _, err := GetSceneFolderURL(goodLandSatID1, badDataType)
	assert.NotNil(t, err, "Invalid scene data type did not cause an error")
	assert.Contains(t, err.Error(), "Unknown LandSat data type")

	_, _, err = GetSceneFolderURL(goodLandSatID1, "")
	assert.NotNil(t, err, "Invalid scene data type did not cause an error")
	assert.Contains(t, err.Error(), "Unknown LandSat data type")
}

func TestGetSceneFolderURL_L1TSceneID(t *testing.T) {
	url, prefix, err := GetSceneFolderURL(goodLandSatID1, l1tDataType)
	host := util.GetLandsatHost()
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, fmt.Sprintf(preCollectionLandSatAWSURL, host, "123", "456", goodLandSatID1, ""), url)
	assert.Equal(t, goodLandSatID1, prefix)
}

func TestGetSceneFolderURL_L1TPSceneID(t *testing.T) {
	UpdateSceneMap(mockLogContext{})
	url, prefix, err := GetSceneFolderURL(goodLandSatID1, l1tpDataType)
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, "https://s3-us-west-2.fakeamazonaws.dummy/thisiscorrect/", url)
	assert.Equal(t, collection1ID1, prefix)
}

func TestGetSceneFolderURL_L1GTInSceneMap(t *testing.T) {
	// NOTE: L1GT data type can mean both Collection-1 and pre-collection 1; this tests the former
	UpdateSceneMap(mockLogContext{})
	url, prefix, err := GetSceneFolderURL(goodLandSatID2, l1gtDataType)
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, "https://s3-us-west-2.fakeamazonaws.dummy/thisisalsocorrect/", url)
	assert.Equal(t, collection1ID2, prefix)
}

func TestGetSceneFolderURL_L1GTNotInSceneMap(t *testing.T) {
	// NOTE: L1GT data type can mean both Collection-1 and pre-collection 1; this tests the latter
	UpdateSceneMap(mockLogContext{})
	url, prefix, err := GetSceneFolderURL(goodLandSatIDNotInSceneMap, l1gtDataType)
	host := util.GetLandsatHost()
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, fmt.Sprintf(preCollectionLandSatAWSURL, host, "102", "938", goodLandSatIDNotInSceneMap, ""), url)
	assert.Equal(t, goodLandSatIDNotInSceneMap, prefix)
}

func TestUpdateSceneMapAsync_Success(t *testing.T) {
	done, errored := UpdateSceneMapAsync(mockLogContext{})
	select {
	case <-done:
		return
	case err := <-errored:
		assert.Fail(t, err.Error())
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timed out")
	}
}

func TestUpdateSceneMapOnTicker(t *testing.T) {
	go UpdateSceneMapOnTicker(500*time.Millisecond, mockLogContext{})

	<-time.After(100 * time.Millisecond)
	assert.True(t, SceneMapIsReady, "Scene map not ready immediately after scene map ticker update")

	SceneMapIsReady = false
	<-time.After(600 * time.Millisecond)
	assert.True(t, SceneMapIsReady, "Scene map not ready again after ticker should have gone off")
}

type mockLogContext struct{}

func (ctx mockLogContext) AppName() string    { return "bf-ia-broker TESTING" }
func (ctx mockLogContext) SessionID() string  { return "test-session" }
func (ctx mockLogContext) LogRootDir() string { return "/tmp" }
