package komodo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/sendgrid/martini"
)

const healthcheckVersion = 1

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
}

// Our HTTP server used to serve http status endpoints
type Server struct {
	// The service that is being monitored
	Service Adminable
	// *Optional* interface to enable and disable debugging
	Debuggable Debuggable

	routes []martini.Route

	// Martini and Router are embedded to make it easier to add HTTP endpoints
	*martini.Martini
	martini.Router
}

// Creates a new instance of a server
//
// An `Adminable` is required
func NewServer(adminable Adminable) *Server {
	router := martini.NewRouter()
	m := martini.New()
	m.Action(router.Handle)

	server := &Server{
		Service: adminable,
		routes:  make([]martini.Route, 0),
		Martini: m,
		Router:  router,
	}

	return server
}

type supportedRoute struct {
	Url string `json:"url"`
}

// HTTP endpoint for reporting the supported routes
//
// Status Codes:
//     - 200: Service is healthy
//     - 500: Some error occurred during the route
//
// Example Response:
//
//     HTTP/1.1 200 Ok
//     Content-Type: application/json
//
//     [
//       {"url": "/"},
//       {"url": "/healthcheck"}
//     ]
//
func (s *Server) Index(routes martini.Routes) (int, []byte) {
	results := make([]supportedRoute, 0, len(s.routes))

	for _, route := range s.routes {
		r := supportedRoute{Url: route.URLWith([]string{})}
		results = append(results, r)
	}

	response, _ := json.Marshal(results)

	return 200, response
}

// Response returned in a healthcheck
type HealthcheckResponse struct {
	AppName            string                       `json:"app_name"`
	AppVersion         string                       `json:"app_version"`
	HealthcheckVersion int                          `json:"healthcheck_version"`
	Host               string                       `json:"host"`
	Results            map[string]HealthcheckResult `json:"results"`
}

// Nested healthcheck result object
type HealthcheckResult struct {
	Name    string  `json:"-"`
	Ok      bool    `json:"ok"`
	Message *string `json:"message"` // Could be nil
}

// HTTP endpoint for handling healthchecks
//
// Services provide a list of healthchecks, which are run concurrently
// in goroutines. The results are aggregated and returned to the client.
//
// Follows the healthcheck specs from:
// https://wiki.sendgrid.net/pages/viewpage.action?pageId=8295814#MonitoringDesign-Healthcheck
//
// Status Codes:
//     - Always returns 200
//
// Example Response:
//
//     HTTP/1.1 200 OK
//     Content-Type: application/json
//
//     {
//         "app_name": "example",
//         "app_version": "0.0.1",
//         "healthcheck_version": 1,
//         "host": "example-001.sjc1.sendgrid.net",
//         "results": {
//             "apid": {
//                 "ok": true,
//                 "message": null
//             },
//             "spool_dir": {
//                 "ok": false,
//                 "message": "Error: No such file /tmp/filterd_recv"
//             }
//         }
//      }
//
func (s *Server) Healthcheck() (int, string) {
	response := &HealthcheckResponse{
		AppName:            s.Service.Name(),
		AppVersion:         s.Service.Version(),
		HealthcheckVersion: healthcheckVersion,
		Host:               hostname,
	}
	response.Results = make(map[string]HealthcheckResult)

	// run all healthchecks concurrently
	results := make(chan HealthcheckResult, len(s.Service.Healthchecks()))
	for _, h := range s.Service.Healthchecks() {
		go func(healthcheck Healthcheck) {
			// run healthcheck and report to channel

			var message *string
			message = nil
			ok := true

			if err := healthcheck.Check(); err != nil {
				ok = false
				errMessage := err.Error()
				message = &errMessage
			}

			results <- HealthcheckResult{
				Name:    healthcheck.Name(),
				Ok:      ok,
				Message: message,
			}
		}(h)
	}

	for i := 0; i < len(s.Service.Healthchecks()); i++ {
		result := <-results
		response.Results[result.Name] = result
	}

	data, err := json.Marshal(response)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Something went wrong with json processing: %q", err)
	}

	return http.StatusOK, string(data)
}

// HTTP endpoint for returning whether or not a service is in maintenance mode
//
// Status Codes:
//     - 200: Service is not in maintenance mode
//     - 400: Service is in maintenance mode
//     - 500: Service encountered an error reading the maintenance file
//
// Example Response:
//
//     HTTP/1.1 200 OK
//     Content-Type: application/json
//
//     {
//       "message": "Not in maintenance mode"
//     }
//
func (s *Server) MaintenanceMode() (code int, output string) {
	var message string
	filename := s.Service.MaintenanceFile()

	// create the json response on function return
	// using named returns
	defer func() {
		result := map[string]string{"message": message}
		jsonResult, err := json.Marshal(result)

		if err != nil {
			code = http.StatusInternalServerError
			output = fmt.Sprintf("Error encoding json: %s", err)
		}

		output = string(jsonResult)
	}()

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			//return 200, "Not in maintenance mode"
			code = http.StatusOK
			message = "Not in maintenance mode"

			return
		}
	}

	file, err := os.Open(filename)
	if err != nil {
		code = http.StatusInternalServerError
		message = fmt.Sprintf("Error opening maintenance file %q: %s", filename, err)
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		code = http.StatusInternalServerError
		message = fmt.Sprintf("Error reading maintenance file %q: %s", filename, err)
		return
	}

	code = http.StatusBadRequest
	message = string(data)

	return
}

// HTTP endpoint for viewing a service's debug status
//
// Status Codes:
//     - 200: Success
//     - 500: Service encountered an error trying to retrieve the debug status
//
// Example Response:
//
//     HTTP/1.1 200 OK
//     Content-Type: application/json
//
//     {
//       "debug": true
//     }
//
func (s *Server) GetDebug() (int, string) {
	debug := s.Debuggable.Debug()

	result := map[string]bool{"debug": debug}
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Error writing json: %s", err)
	}

	return http.StatusOK, string(jsonResult)
}

// HTTP endpoint for enabling a service's debug
//
// Example Response:
//
//     HTTP/1.1 204 No Content
//
func (s *Server) PutDebug() (int, string) {
	s.Debuggable.SetDebug(true)
	return http.StatusNoContent, ""
}

// HTTP endpoint for disabling a service's debug output
//
// Example Response:
//
//     HTTP/1.1 204 No Content
//
func (s *Server) DeleteDebug() (int, string) {
	s.Debuggable.SetDebug(false)
	return http.StatusNoContent, ""
}

// HTTP endpoint for viewing a service's config values
//
// Status Codes:
//     - 200: Success
//     - 500: Service encountered an error trying to retrieve service's config
// Example Response: XXX TODO: actually set content-type to application/json
//
//     HTTP/1.1 200 OK
//     Content-Type: application/json
//
//     {"SMTPHost":"0.0.0.0","SMTPPort":10027,"KomodoHost":"localhost","KomodoPort":4567,...}
//
func (s *Server) Config() (int, string) {
	var data []byte
	config := s.Service.Config()
	data, err := json.Marshal(config)

	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Error writing json: %s", err)
	}
	return http.StatusOK, string(data)
}

// Creates a listener from the provided string and calls serve
func (s *Server) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(l)
}

// Serves http on the given listener
// Typically called in a goroutine
func (s *Server) Serve(l net.Listener) error {
	s.setupEndpoints()

	return http.Serve(l, s.Martini)
}

func (s *Server) setupEndpoints() {
	var r martini.Route

	r = s.Get("/", s.Index)
	s.routes = append(s.routes, r)

	r = s.Get("/healthcheck", s.Healthcheck)
	s.routes = append(s.routes, r)

	r = s.Get("/maintenance_mode", s.MaintenanceMode)
	s.routes = append(s.routes, r)

	r = s.Get("/config", s.Config)
	s.routes = append(s.routes, r)

	if s.Debuggable != nil {
		r = s.Get("/debug", s.GetDebug)
		s.routes = append(s.routes, r)

		r = s.Put("/debug", s.PutDebug)
		s.routes = append(s.routes, r)

		r = s.Delete("/debug", s.DeleteDebug)
		s.routes = append(s.routes, r)
	}

}

// All healthchecks have a name and a value
type Healthcheck interface {
	Name() string
	Check() error
}

// Simple helper struct that satisfies the `Healthcheck` interface
type BasicHealthcheck struct {
	HealthcheckName string
	Healthcheck     func() error
}

func (b *BasicHealthcheck) Name() string {
	return b.HealthcheckName
}

func (b *BasicHealthcheck) Check() error {
	return b.Healthcheck()
}

// Requirements to monitor a service
type Adminable interface {
	// Name of the running Application.
	Name() string

	// Version of the running Application.
	// It can be any string, so it could include the version of dependencies in an app
	Version() string

	// List of healthchecks to run (they will be run concurrently)
	Healthchecks() []Healthcheck

	// File to check for maintenance mode
	MaintenanceFile() string

	// Returns the config struct from the adminable service
	Config() interface{}
}

// Requirements to support a debuggable server
type Debuggable interface {
	// Get the value of debug
	Debug() bool

	// To enable or disable debug output
	SetDebug(bool)
}
