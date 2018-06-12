package landsatlocalindex

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db"
)

func getThumbURLForSceneID(tx *sql.Tx, sceneID string) (*url.URL, error) {
	scene, err := db.GetSceneByID(tx, sceneID)
	if err != nil {
		return nil, err
	}

	sceneURL, _ := url.Parse(scene.SceneURLString)
	return sceneURL.ResolveReference(&url.URL{Path: getThumbFileName(sceneID)}), nil
}

func getThumbFileName(sceneID string) string {
	// TODO: add support for small thumbs as well
	return fmt.Sprintf("%s_thumb_large.jpg", sceneID)
}
