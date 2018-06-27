package metadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/venicegeo/geojson-go/geojson"
)

// GetLandsatS3SceneMetadata retrieves the MTL data for a given scene from a Landsat S3 URL
func GetLandsatS3SceneMetadata(sceneID string, sceneURL string) (*LandsatSceneMetadata, error) {
	baseURL, err := url.Parse(sceneURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base scene URL: %v", err)
	}

	mtlURL := formatMTLURL(sceneID, baseURL)
	mtl, err := getMTL(mtlURL)
	if err != nil {
		return nil, fmt.Errorf("error retrieving/parsing scene MTL: %v", err)
	}

	pm := mtl.L1MetadataFile.ProductMetadata

	return &LandsatSceneMetadata{
		Bounds: geojson.NewPolygon([][][]float64{[][]float64{
			[]float64{pm.CornerUpperLeftLon, pm.CornerLowerLeftLat},
			[]float64{pm.CornerUpperRightLon, pm.CornerUpperRightLat},
			[]float64{pm.CornerLowerRightLon, pm.CornerLowerRightLat},
			[]float64{pm.CornerLowerLeftLon, pm.CornerLowerLeftLat},
			[]float64{pm.CornerUpperLeftLon, pm.CornerLowerLeftLat},
		}}),
	}, nil
}

func formatMTLURL(sceneID string, baseURL *url.URL) *url.URL {
	mtlJSON, _ := url.Parse(fmt.Sprintf("%s_MTL.json", sceneID))
	return baseURL.ResolveReference(mtlJSON)
}

func getMTL(mtlURL *url.URL) (*sceneMTL, error) {
	resp, err := http.Get(mtlURL.String())
	if err != nil {
		return nil, err
	}

	bodyData, _ := ioutil.ReadAll(resp.Body)
	var mtl sceneMTL

	err = json.Unmarshal(bodyData, &mtl)
	if err != nil {
		return nil, err
	}

	return &mtl, nil

}
