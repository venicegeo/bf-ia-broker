package db

import (
	"database/sql"

	"github.com/venicegeo/geojson-go/geojson"
)

func GetSceneByID(tx *sql.Tx, productID string) (*LandsatLocalIndexScene, error) {
	var boundsBytes []byte
	scene := LandsatLocalIndexScene{}

	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds)
		FROM public.scenes
		WHERE product_id=$1
		LIMIT 1`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &boundsBytes)
	if err != nil {
		return nil, err
	}

	scene.Bounds, err = geojson.PolygonFromBytes(boundsBytes)
	if err != nil {
		return nil, err
	}

	return &scene, nil
}
