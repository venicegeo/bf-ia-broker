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

const notSentinelID = "NOT_SENTINEL"
const malformedSentinelID = "S2A_ABCDEF"
const goodSentinelID = "S2A_MSIL1C_20160513T183921_N0204_R070_T11SKD_20160513T185132"

var goodSentinelIDExamples = []string{
	"S2A_MSIL1C_20161208T184752_N0204_R070_T11SKC_20161208T184750",
	"S2A_MSIL1C_20151005T185006_N0204_R070_T10SGH_20161214T094840",
	"S2A_MSIL1C_20151005T185006_N0204_R070_T11SKC_20161214T094840",
	"S2A_MSIL1C_20161221T185802_N0204_R113_T11SKC_20161221T185803",
	"S2A_MSIL1C_20161221T185802_N0204_R113_T10SGH_20161221T185803",
	"S2A_MSIL1C_20170107T184741_N0204_R070_T10SGH_20170107T184740",
	"S2A_MSIL1C_20170107T184741_N0204_R070_T11SKC_20170107T184740",
}

func TestSentinelPrefixDetection(t *testing.T) {
	assert.True(t, isSentinelFeature(goodSentinelID))
	assert.True(t, isSentinelFeature(malformedSentinelID))
	assert.False(t, isSentinelFeature(notSentinelID))
}

func TestSentinelIDRegex(t *testing.T) {
	for _, id := range goodSentinelIDExamples {
		assert.True(t, sentinelIDPattern.MatchString(id))
	}
}

func TestAddSentinelBands_NoOpWhenNotSentinel(t *testing.T) {
	assert.Nil(t, addSentinelS3BandsToProperties(notSentinelID, &map[string]interface{}{}))
}

func TestAddSentinelBands_ErrorWhenMalformedID(t *testing.T) {
	assert.NotNil(t, addSentinelS3BandsToProperties(malformedSentinelID, &map[string]interface{}{}))
}

func TestAddSentinelBands(t *testing.T) {
	properties := map[string]interface{}{}
	err := addSentinelS3BandsToProperties(goodSentinelID, &properties)
	assert.Nil(t, err)

	bands, ok := properties["bands"]
	assert.True(t, ok, "missing 'bands' in properties")

	bandsMap := bands.(map[string]string)
	for band, filename := range sentinelBandsFilenames {
		url, found := bandsMap[band]
		assert.True(t, found, "missing band: "+band)
		assert.Contains(t, url, "/11/S/KD/", "URL does not contain correct AWS path")
		assert.True(t, strings.HasSuffix(url, filename), "wrong filename for band; GOT: %s EXPECTED: %s", url, filename)
	}
}
