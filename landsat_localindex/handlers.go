package landsatlocalindex

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

// DiscoverHandler is a handler for /localindex/discover/landsat
// @Title localIndexDiscoverHandler
// @Description discovers scenes from Planet Labs
// @Accept  plain
// @Param   bbox            query   string  false        "The bounding box, as a GeoJSON Bounding box (x1,y1,x2,y2)"
// @Param   cloudCover      query   string  false        "The maximum cloud cover, as a percentage (0-100)"
// @Param   acquiredDate    query   string  false        "The minimum (earliest) acquired date, as RFC 3339"
// @Param   maxAcquiredDate query   string  false        "The maximum acquired date, as RFC 3339"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.FeatureCollection
// @Failure 400 {object}  string
// @Router /localindex/discover/{itemType} [get]
type DiscoverHandler struct {
	Context Context
}

// NewDiscoverHandler creates a new handler using configuration
// from environment variables
func NewDiscoverHandler(connectionProvider db.ConnectionProvider) (*DiscoverHandler, error) {
	tidesURL := util.GetTidesURL()

	db, err := connectionProvider(&util.BasicLogContext{})
	if err != nil {
		return nil, err
	}

	return &DiscoverHandler{
		Context: Context{
			DB:           db,
			BaseTidesURL: tidesURL,
		},
	}, nil
}

// ServeHTTP implements the http.Handler interface for the DiscoverHandler type
func (h DiscoverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tx, err := h.Context.DB.Begin()
	if err != nil {
		message := fmt.Sprintf("Could not begin DB transaction: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	tides, _ := strconv.ParseBool(r.FormValue("tides"))
	bbox, err := geojson.NewBoundingBox(r.FormValue("bbox"))
	if err != nil {
		message := fmt.Sprintf("The bbox value of %v is invalid", r.FormValue("bbox"))
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusBadRequest)
		tx.Rollback()
		return
	}
	maxCloudCover := float64(1)
	if r.FormValue("cloudCover") != "" {
		if maxCloudCover, err = strconv.ParseFloat(r.FormValue("cloudCover"), 64); err != nil {
			message := fmt.Sprintf("Cloud Cover value of %v is invalid.", r.FormValue("cloudCover"))
			util.LogSimpleErr(&h.Context, message, err)
			util.HTTPError(r, w, &h.Context, message, http.StatusBadRequest)
			tx.Rollback()
			return
		}
		maxCloudCover = maxCloudCover / 100.0
	}
	minAcquiredDate := time.Unix(0, 0)
	if r.FormValue("acquiredDate") != "" {
		if minAcquiredDate, err = time.Parse(time.RFC3339, r.FormValue("acquiredDate")); err != nil {
			message := fmt.Sprintf("Acquired date value of %v is invalid.", r.FormValue("acquiredDate"))
			util.LogSimpleErr(&h.Context, message, err)
			util.HTTPError(r, w, &h.Context, message, http.StatusBadRequest)
			tx.Rollback()
			return
		}
	}
	maxAcquiredDate := time.Now()
	if r.FormValue("maxAcquiredDate") != "" {
		if maxAcquiredDate, err = time.Parse(time.RFC3339, r.FormValue("maxAcquiredDate")); err != nil {
			message := fmt.Sprintf("Acquired date value of %v is invalid.", r.FormValue("maxAcquiredDate"))
			util.LogSimpleErr(&h.Context, message, err)
			util.HTTPError(r, w, &h.Context, message, http.StatusBadRequest)
			tx.Rollback()
			return
		}
	}

	multiResult, err := discoverScenes(tx, h.Context, bbox, maxCloudCover, minAcquiredDate, maxAcquiredDate, tides)

	if err != nil {
		message := fmt.Sprintf("Error searching for scenes: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	featureCollection, err := multiResult.GeoJSONFeatureCollection()
	if err != nil {
		message := fmt.Sprintf("Error converting to feature collection: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		return
	}
	w.Write([]byte(featureCollection.String()))
}

// MetadataHandler is a handler for /localindex/landsat/{id}
// @Title localIndexMetadataHandler
// @Description discovers scenes from Planet Labs
// @Accept  plain
// @Param   id            path   string  false        "The ID of the requested scene"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.Feature
// @Failure 400 {object}  string
// @Router /localindex/landsat/{id} [get]
type MetadataHandler struct {
	Context Context
}

// NewMetadataHandler creates a new handler using the environment and given DB
func NewMetadataHandler(connectionProvider db.ConnectionProvider) (*MetadataHandler, error) {
	tidesURL := util.GetTidesURL()

	db, err := connectionProvider(&util.BasicLogContext{})
	if err != nil {
		return nil, err
	}

	return &MetadataHandler{
		Context: Context{
			DB:           db,
			BaseTidesURL: tidesURL,
		},
	}, nil
}

func (h MetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sceneID, ok := mux.Vars(r)["id"]
	if !ok {
		message := "No scene ID found in URL"
		util.LogAlert(&h.Context, message)
		util.HTTPError(r, w, &h.Context, message, http.StatusNotFound)
		return
	}

	tx, err := h.Context.DB.Begin()
	if err != nil {
		message := fmt.Sprintf("Could not begin DB transaction: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	tides, _ := strconv.ParseBool(r.FormValue("tides"))

	metadata, err := getMetadata(tx, h.Context, sceneID, tides)
	if err == sql.ErrNoRows {
		message := fmt.Sprintf("Scene not found: %s", sceneID)
		util.LogInfo(&h.Context, message)
		util.HTTPError(r, w, &h.Context, message, http.StatusNotFound)
		tx.Rollback()
		return
	}
	if err != nil {
		message := fmt.Sprintf("Server error searching for scene: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	feature, err := metadata.GeoJSONFeature()
	if err != nil {
		message := fmt.Sprintf("Error converting metadata to geojson: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	w.Write([]byte(feature.String()))
}

// XYZTileHandler is a handler for /localindex/tiles/landsat/{id}/{Z}/{X}/{Y}.jpg
// @Title localIndexXYZTileHandler
// @Description performs a redirect to the correct AWS-hosted map tile
// @Accept  plain
// @Success 302 redirect to actual image
// @Failure 400 {object}  string
// @Router /localindex/tiles/landsat/{id}/{Z}/{X}/Y.jpg [get]
type XYZTileHandler struct {
	Context Context
}

// NewXYZTileHandler creates a new handler using the environment and given DB
func NewXYZTileHandler(connectionProvider db.ConnectionProvider) (*XYZTileHandler, error) {
	db, err := connectionProvider(&util.BasicLogContext{})
	if err != nil {
		return nil, err
	}

	return &XYZTileHandler{
		Context: Context{
			DB: db,
		},
	}, nil
}

func (h XYZTileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sceneID, ok := mux.Vars(r)["id"]
	if !ok {
		message := "No scene ID found in URL"
		util.LogAlert(&h.Context, message)
		util.HTTPError(r, w, &h.Context, message, http.StatusNotFound)
		return
	}

	tx, err := h.Context.DB.Begin()
	if err != nil {
		message := fmt.Sprintf("Could not begin DB transaction: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	thumbURL, err := getThumbURLForSceneID(tx, sceneID)
	if err == sql.ErrNoRows {
		message := fmt.Sprintf("Scene not found: %s", sceneID)
		util.LogInfo(&h.Context, message)
		util.HTTPError(r, w, &h.Context, message, http.StatusNotFound)
		tx.Rollback()
		return
	}
	if err != nil {
		message := fmt.Sprintf("Server error searching for scene: %v", err)
		util.LogSimpleErr(&h.Context, message, err)
		util.HTTPError(r, w, &h.Context, message, http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	w.Header().Set("Location", thumbURL.String())
	w.WriteHeader(http.StatusFound)
}
