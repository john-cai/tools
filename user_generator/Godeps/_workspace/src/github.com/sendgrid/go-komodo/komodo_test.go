package komodo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

var version = fmt.Sprintf("komodo: %s; app: %s", VERSION, "0.1.0-alpha")
var maintenanceString = "Maintenance File"

type FakeAdminable struct {
	debug        bool
	healthchecks []Healthcheck
}

type FakeConfig struct {
	Fake        bool
	Description string
	Length      int
}

var MyFakeConfig = FakeConfig{
	Fake:        true,
	Description: "I'm a big, fat phony!",
	Length:      20}

func (f *FakeAdminable) Version() string {
	return version
}
func (f *FakeAdminable) Name() string {
	return "fake_adminable"
}

func (f *FakeAdminable) Healthchecks() []Healthcheck {
	if f.healthchecks == nil {
		return []Healthcheck{}
	}
	return f.healthchecks
}

func (f *FakeAdminable) MaintenanceFile() string {
	file, _ := ioutil.TempFile("/tmp", "go-komodo")
	file.Write([]byte(maintenanceString))
	file.Close()
	return file.Name()
}

func (f *FakeAdminable) SetDebug(d bool) {
	f.debug = d
}

func (f *FakeAdminable) Debug() bool {
	return f.debug
}

func (f *FakeAdminable) Config() interface{} {
	return MyFakeConfig
}

func TestListenAndServe(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("Unable to start listening", err)
	}

	fa := &FakeAdminable{}
	server := NewServer(fa)

	go func() {
		err := server.Serve(l)
		if err != nil {
			t.Fatal("Unable to serve", err)
		}
	}()

	resp, err := http.Get(fmt.Sprintf("http://%s/healthcheck", l.Addr().String()))
	if err != nil {
		t.Error("Unable to get healthcheck", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200 for healthcheck, got %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response from healthcheck: %s", err)
	}

	jsonData := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		t.Errorf("Healthcheck response is not json: %s", err)
	}

	if jsonData["app_version"] != fa.Version() {
		t.Errorf("Error with version in healthcheck; expected %q, got %q", fa.Version(), jsonData["app_version"])
	}

	if jsonData["app_name"] != fa.Name() {
		t.Errorf("Error with name in healthcheck; expected %q, got %q", fa.Name(), jsonData["app_name"])
	}

	if _, ok := jsonData["results"]; !ok {
		t.Errorf(`healthcheck "results" key is missing`)
	}
}

func TestIndexRoute(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("Unable to start listening", err)
	}

	server := NewServer(&FakeAdminable{})

	go func() {
		err := server.Serve(l)
		if err != nil {
			t.Fatal("Unable to serve", err)
		}
	}()

	resp, err := http.Get(fmt.Sprintf("http://%s/", l.Addr().String()))
	if err != nil {
		t.Fatal("Error accessing index route:", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Response status code was incorrect, got %d", resp.StatusCode)
	}

	expectedRoutes := len(server.routes)

	rawData, _ := ioutil.ReadAll(resp.Body)
	data := make([]interface{}, 0)
	err = json.Unmarshal(rawData, &data)

	if err != nil {
		t.Errorf("Error unmarshalling the json: %s", err)
	}

	if len(data) != expectedRoutes {
		t.Errorf("Got unexpected number of routes in response: %d", len(data))
	}
}

func TestMaintenanceMode(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("Unable to start listening", err)
	}

	server := NewServer(&FakeAdminable{})

	go func() {
		err := server.Serve(l)
		if err != nil {
			t.Fatal("Unable to serve", err)
		}
	}()

	resp, err := http.Get(fmt.Sprintf("http://%s/maintenance_mode", l.Addr().String()))
	if resp.StatusCode != 400 {
		t.Errorf("Expected status code 400 for maintenance_mode, got %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response from healthcheck: %s", err)
	}

	jsonData := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		t.Errorf("Healthcheck response is not json: %s", err)
	}

	if jsonData["message"] != maintenanceString {
		t.Errorf("Error with maintenance file response; expected %q, got %q", maintenanceString, jsonData["message"])
	}
}

func TestDebuggable(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("Unable to start listening", err)
	}

	adminable := &FakeAdminable{}
	server := NewServer(adminable)
	server.Debuggable = adminable

	go func() {
		err := server.Serve(l)
		if err != nil {
			t.Fatal("Unable to serve", err)
		}
	}()

	client := http.Client{}

	url := fmt.Sprintf("http://%s/debug", l.Addr().String())
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		t.Fatal("Unable to create request: ", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Error("Unable to PUT debug", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	if !adminable.Debug() {
		t.Error("Debug flag was not set")
	}

	resp, err = http.Get(url)
	if err != nil {
		t.Error("Unable to GET debug", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Error reading response", err)
	}
	resp.Body.Close()

	result := make(map[string]bool)
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Error("Invalid json received", err)
	}

	debug, ok := result["debug"]
	if !debug {
		t.Errorf("Got incorrect value for debug, expected %t got %t", true, debug)
	}
	if !ok {
		t.Error("JSON did not contain 'debug' key")
	}

	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatal("Unable to create request: ", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Error("Unable to DELETE debug", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	if adminable.Debug() {
		t.Error("Debug flag was not unset")
	}

	resp, err = http.Get(url)
	if err != nil {
		t.Error("Unable to GET debug", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Error reading response", err)
	}
	resp.Body.Close()

	result = make(map[string]bool)
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Error("Invalid json received", err)
	}

	debug, ok = result["debug"]
	if debug {
		t.Errorf("Got incorrect value for debug, expected %v got %v", debug, false)
	}
	if !ok {
		t.Error("JSON did not contain 'debug' key")
	}
}

func TestConfigable(t *testing.T) {
	server := NewServer(&FakeAdminable{})

	statusCode, configData := server.Config()
	if statusCode != 200 {
		t.Fatal("Non 200 response from Config endpoint", statusCode)
	}

	fromResponse := &FakeConfig{}
	err := json.Unmarshal([]byte(configData), fromResponse)
	if err != nil {
		t.Error("Unable to repopulate config struct with data from config endpoint", err)
	}

	if *fromResponse != MyFakeConfig {
		t.Errorf("Config mutated! Got %+v, expected %+v", fromResponse, MyFakeConfig)
	}
}

func TestHealthChecks(t *testing.T) {
	l, err := net.Listen("tcp", ":4567")
	if err != nil {
		t.Fatal("Unable to start listening", err)
	}

	fa := &FakeAdminable{healthchecks: make([]Healthcheck, 0)}

	expectedResults := map[string]string{"health_check_1": "health_check_1_error", "health_check_2": "health_check_2_error"}

	for k, v := range expectedResults {
		// need to scope v because it's used inside of a closure
		e := v
		fa.healthchecks = append(fa.healthchecks, &BasicHealthcheck{HealthcheckName: k, Healthcheck: func() error { return errors.New(e) }})
	}

	server := NewServer(fa)

	go func() {
		err := server.Serve(l)
		if err != nil {
			t.Fatal("Unable to serve", err)
		}
	}()

	resp, err := http.Get(fmt.Sprintf("http://%s/healthcheck", l.Addr().String()))
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200 for healthchecks, got %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response from healthcheck: %s", err)
	}

	healthCheckResponse := &HealthcheckResponse{}

	err = json.Unmarshal(data, healthCheckResponse)

	if err != nil {
		t.Errorf("Could not unmarshall %s", err)
	}

	for healthcheckName, result := range healthCheckResponse.Results {
		if *result.Message != expectedResults[healthcheckName] {
			t.Errorf("Expected error %s, got %s.", expectedResults[healthcheckName], *result.Message)
		}
	}
}
