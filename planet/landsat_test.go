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

package planet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const notLandSatID = "NOT_LANDSAT"
const malformedLandSatID = "LC8ABC"
const goodLandSatID = "LC80060522017107LGN00"
const l1tDataType = "L1T"

func TestAddLandSatBands_ErrorWhenMalformedID(t *testing.T) {
	err := addLandsatS3BandsToProperties(malformedLandSatID, l1tDataType, &map[string]interface{}{})
	assert.NotNil(t, err)
}

func TestAddLandSatBands(t *testing.T) {
	properties := map[string]interface{}{}
	err := addLandsatS3BandsToProperties(goodLandSatID, l1tDataType, &properties)
	assert.Nil(t, err)

	bands, ok := properties["bands"]
	assert.True(t, ok, "missing 'bands' in properties")

	bandsMap := bands.(map[string]string)
	for band, suffix := range landSatBandsSuffixes {
		url, found := bandsMap[band]
		assert.True(t, found, "missing band: "+band)
		assert.Contains(t, url, "/006/052/", "URL does not contain correct AWS path")
		assert.Contains(t, url, goodLandSatID)
		assert.True(t, strings.HasSuffix(url, suffix), "wrong suffix for band")
	}
}
