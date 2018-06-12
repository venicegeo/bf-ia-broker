package planet

import (
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

var disablePermissionsCheck bool

func init() {
	disablePermissionsCheck, _ = util.IsPlanetPermissionsDisabled()
	if disablePermissionsCheck {
		util.LogInfo(&util.BasicLogContext{}, "Disabling Planet Labs permissions check")
	}
}

// Context is the context for a Planet Labs Operation
type Context struct {
	BasePlanetURL string
	BaseTidesURL  string
	PlanetKey     string
	sessionID     string
}

// AppName returns an empty string
func (c *Context) AppName() string {
	return "bf-ia-broker"
}

// SessionID returns a Session ID, creating one if needed
func (c *Context) SessionID() string {
	if c.sessionID == "" {
		c.sessionID, _ = util.PsuUUID()
	}
	return c.sessionID
}

// LogRootDir returns an empty string
func (c *Context) LogRootDir() string {
	return ""
}

// SearchOptions are the search options for a quick-search request
type SearchOptions struct {
	ItemType        string
	Tides           bool
	AcquiredDate    string
	MaxAcquiredDate string
	Bbox            geojson.BoundingBox
	CloudCover      float64
}

type searchResults struct {
	Features []feature `json:"features"`
}

type feature struct {
	Links       Links    `json:"_links"`
	Permissions []string `json:"_permissions"`
}

type request struct {
	ItemTypes []string `json:"item_types"`
	Filter    filter   `json:"filter"`
}

type filter struct {
	Type   string        `json:"type"`
	Config []interface{} `json:"config"`
}

type objectFilter struct {
	Type      string      `json:"type"`
	FieldName string      `json:"field_name"`
	Config    interface{} `json:"config"`
}

type dateConfig struct {
	GTE string `json:"gte,omitempty"`
	LTE string `json:"lte,omitempty"`
	GT  string `json:"gt,omitempty"`
	LT  string `json:"lt,omitempty"`
}

type rangeConfig struct {
	GTE float64 `json:"gte,omitempty"`
	LTE float64 `json:"lte,omitempty"`
	GT  float64 `json:"gt,omitempty"`
	LT  float64 `json:"lt,omitempty"`
}

// Assets represents the assets available for a scene
type Assets struct {
	Analytic    Asset `json:"analytic"`
	AnalyticXML Asset `json:"analytic_xml"`
	UDM         Asset `json:"udm"`
	Visual      Asset `json:"visual"`
	VisualXML   Asset `json:"visual_xml"`
}

// Asset represents a single asset available for a scene
type Asset struct {
	Links       Links    `json:"_links"`
	Status      string   `json:"status"`
	Type        string   `json:"type"`
	Location    string   `json:"location,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	Permissions []string `json:"_permissions,omitempty"`
}

// Links represents the links JSON structure.
type Links struct {
	Self     string `json:"_self"`
	Activate string `json:"activate"`
	Type     string `json:"type"`
}

type planetRequestInput struct {
	method      string
	inputURL    string // URL may be relative or absolute based on baseURLString
	body        []byte
	contentType string
}

// MetadataOptions are the options for the Asset func
type MetadataOptions struct {
	ID       string
	Tides    bool
	ItemType string
}
