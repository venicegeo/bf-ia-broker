package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

// GetSceneByID looks up a single scene by its product ID
func GetSceneByID(tx *sql.Tx, productID string) (*LandsatLocalIndexScene, error) {
	var (
		mtlBoundsBytes     []byte
		mtlBounds          SingleOrMultiPolygon
		polyErr1, polyErr2 error
	)
	scene := LandsatLocalIndexScene{}

	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds)
		FROM public.scenes
		WHERE product_id=$1
			AND corner_ll IS NOT NULL 
		LIMIT 1`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &mtlBoundsBytes)
	if err != nil {
		return nil, err
	}

	if mtlBounds, polyErr1 = geojson.PolygonFromBytes(mtlBoundsBytes); polyErr1 != nil {
		mtlBounds, polyErr2 = geojson.MultiPolygonFromBytes(mtlBoundsBytes)
	}
	if polyErr2 != nil {
		return nil, fmt.Errorf("Could not extract either Polygon or MultiPolygon from MTL bounds bytes: %v; %v", polyErr1, polyErr2)
	}

	scene.Bounds = mtlBounds
	scene.BoundingBox = mtlBounds.ForceBbox()

	return &scene, nil
}

// SearchScenes does a lookup in indexed scenes based on a bounding box, cloud cover, and time window
// Note: Any cloud cover that is <0 is usually corrupt in some way, and shall be excluded
func SearchScenes(tx *sql.Tx, bbox geojson.BoundingBox, maxCloudCover float64, minAcquiredDate time.Time, maxAcquiredDate time.Time) ([]LandsatLocalIndexScene, error) {
	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds)
		FROM public.scenes
		WHERE cloud_cover >= 0
			AND cloud_cover < $1
			AND acquisition_date > $2
			AND acquisition_date < $3
			AND corner_ll IS NOT NULL 
			AND ST_Intersects(bounds, ST_MakeEnvelope($4, $5, $6, $7, 4326))
		ORDER BY acquisition_date DESC
		LIMIT 100`,
		maxCloudCover*100, // Cloud cover is imported as 0-100, not as 0-1
		minAcquiredDate, maxAcquiredDate,
		bbox[0], bbox[1], bbox[2], bbox[3],
	)
	if err != nil {
		return nil, err
	}

	results := []LandsatLocalIndexScene{}
	for rows.Next() {
		var (
			mtlBoundsBytes     []byte
			polyErr1, polyErr2 error
			mtlBounds          SingleOrMultiPolygon
		)

		scene := LandsatLocalIndexScene{}
		if err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &mtlBoundsBytes); err != nil {
			return nil, err
		}

		if mtlBounds, polyErr1 = geojson.PolygonFromBytes(mtlBoundsBytes); polyErr1 != nil {
			mtlBounds, polyErr2 = geojson.MultiPolygonFromBytes(mtlBoundsBytes)
		}
		if polyErr2 != nil {
			return nil, fmt.Errorf("Could not extract either Polygon or MultiPolygon from MTL bounds bytes: %v; %v", polyErr1, polyErr2)
		}

		scene.Bounds = mtlBounds
		scene.BoundingBox = mtlBounds.ForceBbox()

		results = append(results, scene)
	}

	return results, nil
}
