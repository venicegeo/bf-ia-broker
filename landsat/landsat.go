package landsat

import (
	"fmt"
	"regexp"
	"strings"
)

// Old LandSat IDs come back in the form LC80060522017107LGN00

var landSatSceneIDPattern = regexp.MustCompile("LC8([0-9]{3})([0-9]{3}).*")

// IsValidLandSatID returns whether an ID is a valid LandSat ID
func IsValidLandSatID(sceneID string) bool {
	return landSatSceneIDPattern.MatchString(sceneID)
}

const preCollectionLandSatAWSURL = "https://landsat-pds.s3.amazonaws.com/L8/%s/%s/%s/%s"

func formatPreCollectionIDToURL(sceneID string) string {
	m := landSatSceneIDPattern.FindStringSubmatch(sceneID)[1:]
	return fmt.Sprintf(preCollectionLandSatAWSURL, m[0], m[1], sceneID, "")
}

var preCollectionDataTypes = []string{"L1T", "L1GT", "L1G"}

// IsPreCollectionDataType returns whether a data type is a Pre-"Collection 1" type
// Reference: https://landsat.usgs.gov/landsat-processing-details
func IsPreCollectionDataType(dataType string) bool {
	for _, t := range preCollectionDataTypes {
		if dataType == t {
			return true
		}
	}
	return false
}

var collection1DataTypes = []string{"L1TP"}

// IsCollection1DataType returns whether a data type is a "Collection 1" type
// Reference: https://landsat.usgs.gov/landsat-processing-details
func IsCollection1DataType(dataType string) bool {
	dataType = strings.ToUpper(dataType)
	for _, t := range collection1DataTypes {
		if dataType == t {
			return true
		}
	}
	return false
}
