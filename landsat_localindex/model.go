package landsatlocalindex

import (
	"database/sql"

	"github.com/venicegeo/bf-ia-broker/util"
)

// Context is the context for a Planet Labs Operation
type Context struct {
	DB           *sql.DB
	BaseTidesURL string
	sessionID    string
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
