package planet

import (
	"fmt"
	"net/url"
	"time"

	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

func parseSearchResults(context *Context, body []byte) ([]model.BrokerSearchResult, error) {
	planetFeatureCollection, err := planetRawBytesToFeatureCollection(context, body)
	if err != nil {
		return nil, err
	}

	results := make([]model.BrokerSearchResult, len(planetFeatureCollection.Features))
	for i, feature := range planetFeatureCollection.Features {
		result, err := planetSearchBrokerResultFromFeature(feature)
		if err != nil {
			return nil, err
		}
		results[i] = *result
	}

	return results, nil
}

func planetRawBytesToFeatureCollection(context *Context, body []byte) (*geojson.FeatureCollection, error) {
	var (
		planetFeatureCollection *geojson.FeatureCollection
		geoJSONParsedData       interface{}
		ok                      bool
		err                     error
	)
	if geoJSONParsedData, err = geojson.Parse(body); err != nil {
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to parse GeoJSON.\n%v", string(body)), err)
		return nil, err
	}

	if planetFeatureCollection, ok = geoJSONParsedData.(*geojson.FeatureCollection); !ok {
		plErr := util.Error{SimpleMsg: fmt.Sprintf("Expected a FeatureCollection and got %T", geoJSONParsedData), Response: string(body)}
		err = plErr.Log(context, "")
		return nil, err
	}

	return planetFeatureCollection, nil
}

func basicBrokerResultFromPlanetFeature(feature *geojson.Feature, fileFormat model.BrokerFileFormat) (*model.BasicBrokerResult, error) {
	acquiredDate, err := model.ParsePlanetTime(feature.PropertyString("acquired"))
	if err != nil {
		return nil, err
	}
	cloudCover := -1.0
	if _, ok := feature.Properties["cloud_cover"]; ok {
		cloudCover = feature.PropertyFloat("cloud_cover") * 100
	}

	return &model.BasicBrokerResult{
		AcquiredDate: acquiredDate,
		CloudCover:   cloudCover,
		FileFormat:   fileFormat,
		Geometry:     feature.Geometry,
		ID:           feature.IDStr(),
		Resolution:   feature.PropertyFloat("gsd"),
		SensorName:   feature.PropertyString("satellite_id"),
		DataType:     feature.PropertyString("data_type"),
	}, nil
}

func planetSearchBrokerResultFromFeature(feature *geojson.Feature) (*model.BrokerSearchResult, error) {
	// NOTE: all Planet results are GeoTIFF, hence the GeoTIFF hardcoding
	basicBrokerResult, err := basicBrokerResultFromPlanetFeature(feature, model.GeoTIFF)

	if err != nil {
		return nil, err
	}

	return &model.BrokerSearchResult{BasicBrokerResult: *basicBrokerResult}, nil
}

// planetAssetMetadataFromAssets constructs a PlanetAssetMetadata by extracting
// data from a planet.Assets response container
func planetAssetMetadataFromAssets(assets Assets) (*model.PlanetAssetMetadata, error) {
	var err error
	if assets.Analytic.Type == "" {
		// No data means just return nil
		return nil, nil
	}

	expiresAt := time.Time{}
	if assets.Analytic.ExpiresAt != "" {
		expiresAt, err = model.ParsePlanetTime(assets.Analytic.ExpiresAt)
		if err != nil {
			return nil, err
		}
	}
	permissionsCopy := append([]string{}, assets.Analytic.Permissions...)

	assetURL, err := url.Parse(assets.Analytic.Location)
	if err != nil {
		return nil, err
	}
	activationURL, err := url.Parse(assets.Analytic.Links.Activate)
	if err != nil {
		return nil, err
	}

	return &model.PlanetAssetMetadata{
		AssetURL:      *assetURL,
		ActivationURL: *activationURL,
		ExpiresAt:     expiresAt,
		Permissions:   permissionsCopy,
		Status:        assets.Analytic.Status,
		Type:          assets.Analytic.Type,
	}, nil
}
