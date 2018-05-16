package landsataws

import (
	"github.com/gorilla/mux"
)

//NewRouter creates a new router.
func NewRouter() *mux.Router {
	router := mux.NewRouter()

	return router
}
