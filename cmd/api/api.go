package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/healthcheck-watchdog/cmd/api/controller"
	"github.com/healthcheck-watchdog/cmd/manager"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler Handler
}

type Routes []Route

var api controller.ApiController

func NewRouter(manager *manager.Manager) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var routerHandler http.Handler
		routerHandler = route.Handler
		routerHandler = loggerHandler(routerHandler, route.Name)

		router.
			Methods(route.Method).
			Path("/" + route.Pattern).
			Name(route.Name).
			Handler(routerHandler)
	}

	router.
		Methods("GET").
		Path("/metrics").
		Name("Metrics").
		Handler(promhttp.Handler())

	router.
		Methods("GET").
		Path("/probe").
		Name("Metrics").
		Handler(promhttp.Handler())

	api = controller.ApiController{Manager: manager}

	log.Info(
		fmt.Sprintf("API Server initialized on route http://localhost:2112/%s...", "metrics"))
	log.Info(
		fmt.Sprintf("Prometheus Server initialized on route http://localhost:2112/%s...", "probe"))

	return router
}

func loggerHandler(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)
		log.Info(fmt.Sprintf("%s %s %s %s", r.Method, r.RequestURI, name, time.Since(start)))
	})
}

var routes = Routes{
	// swagger:operation GET /ping Health Ping
	// ---
	// summary: Health API
	// description: Returns pong
	// responses:
	//   "200":
	//     description: "pong"
	//     schema: {
	//		"type": "string",
	//	   }
	//   "401":
	//     description: "Unauthorized"
	//   "403":
	//     description: "Forbidden"
	//   "500":
	//     description: "Internal server error"
	Route{
		"Ping",
		"GET",
		"ping",
		Handler{H: api.Ping},
	},

	// swagger:operation GET /health Health Health
	// ---
	// summary: Health API
	// description: Returns pong
	// responses:
	//   "200":
	//     description: "pong"
	//     schema: {
	//		"type": "string",
	//	   }
	//   "401":
	//     description: "Unauthorized"
	//   "403":
	//     description: "Forbidden"
	//   "500":
	//     description: "Internal server error"
	Route{
		"Health",
		"GET",
		"health",
		Handler{H: api.Health},
	},
}
