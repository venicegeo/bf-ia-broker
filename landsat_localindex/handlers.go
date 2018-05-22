package landsatlocalindex

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db"
	"github.com/venicegeo/bf-ia-broker/util"
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
		util.HTTPError(r, w, &h.Context, "Could not begin DB transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	scene, err := db.GetSceneByID(tx, "LC08_L1TP_012029_20170411_20170415_01_T1")
	if err == sql.ErrNoRows {
		util.HTTPError(r, w, &h.Context, "Scene not found", http.StatusNotFound)
		return
	}
	if err != nil {
		util.HTTPError(r, w, &h.Context, "DB error searching for scene: "+err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	w.Write([]byte(scene.SceneURLString))
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
		util.HTTPError(r, w, &h.Context, "Scene not found", http.StatusNotFound)
		return
	}

	tx, err := h.Context.DB.Begin()
	if err != nil {
		util.HTTPError(r, w, &h.Context, "Could not begin DB transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	tides, _ := strconv.ParseBool(r.FormValue("tides"))

	metadata, err := getMetadata(tx, h.Context, sceneID, tides)
	if err == sql.ErrNoRows {
		util.HTTPError(r, w, &h.Context, "Scene not found", http.StatusNotFound)
		return
	}
	if err != nil {
		util.HTTPError(r, w, &h.Context, "DB error searching for scene: "+err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	feature, err := metadata.GeoJSONFeature()
	if err != nil {
		util.HTTPError(r, w, &h.Context, "Error converting metadata to GeoJSON: "+err.Error(), http.StatusInternalServerError)
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
		util.HTTPError(r, w, &h.Context, "Scene not found", http.StatusNotFound)
		return
	}

	tx, err := h.Context.DB.Begin()
	if err != nil {
		util.HTTPError(r, w, &h.Context, "Could not begin DB transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	thumbURL, err := getThumbURLForSceneID(tx, sceneID)
	if err == sql.ErrNoRows {
		util.HTTPError(r, w, &h.Context, "Scene not found", http.StatusNotFound)
		return
	}
	if err != nil {
		util.HTTPError(r, w, &h.Context, "DB error searching for scene: "+err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	w.Header().Set("Location", thumbURL.String())
	w.WriteHeader(http.StatusFound)
}
