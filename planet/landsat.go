package planet

import (
	"errors"

	"github.com/venicegeo/bf-ia-broker/landsat"
)

var landSatBandsSuffixes = map[string]string{
	"coastal":      "_B1.TIF",
	"blue":         "_B2.TIF",
	"green":        "_B3.TIF",
	"red":          "_B4.TIF",
	"nir":          "_B5.TIF",
	"swir1":        "_B6.TIF",
	"swir2":        "_B7.TIF",
	"panchromatic": "_B8.TIF",
	"cirrus":       "_B9.TIF",
	"tirs1":        "_B10.TIF",
	"tirs2":        "_B11.TIF",
}

func addLandsatS3BandsToProperties(landSatID string, dataType string, properties *map[string]interface{}) error {
	if !landsat.IsValidLandSatID(landSatID) {
		return errors.New("Not a valid LandSat ID: " + landSatID)
	}

	awsFolder, prefix, err := landsat.GetSceneFolderURL(landSatID, dataType)
	if err != nil {
		return err
	}

	bands := make(map[string]string)
	for band, suffix := range landSatBandsSuffixes {
		bands[band] = awsFolder + prefix + suffix
	}
	(*properties)["bands"] = bands

	return nil
}
