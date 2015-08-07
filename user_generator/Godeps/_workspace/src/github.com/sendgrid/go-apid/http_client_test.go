package apid

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func functions() map[string]map[string]APIdFunction {
	functions := make(map[string]map[string]APIdFunction)
	functions["transactional"] = map[string]APIdFunction{
		"addTemplate":         "addTeslaTemplate",
		"getTemplate":         "getTeslaTemplate",
		"editTemplate":        "editTeslaTemplate",
		"deleteTemplate":      "deleteTeslaTemplate",
		"getUserTemplateList": "getAllTeslaTemplates",

		"getVersionsForTemplate":      "getAllTeslaVersionsForTemplate",
		"getActiveVersionForTemplate": "getTeslaActiveVersion",
		"getVersionCount":             "getTeslaVersionsCount",

		"addVersion":    "addTeslaVersion",
		"getVersion":    "getTeslaVersion",
		"deleteVersion": "deleteTeslaVersion",
		"editVersion":   "editTeslaVersion",
	}
	functions["marketing"] = map[string]APIdFunction{
		"addVersion":            "addMarketingTemplate",
		"getVersion":            "getMarketingTemplate",
		"deleteVersion":         "deleteMarketingTemplate",
		"getUserVersionList":    "getMarketingTemplates",
		"editVersion":           "editMarketingTemplate",
		"getGlobalTemplateList": "getMarketingSystemTemplates",
	}
	functions["system"] = map[string]APIdFunction{
		"healthCheck": "alive",
	}
	return functions
}

const functionsListJSON string = `
	{
		"functions": {
			"getCredentials": {
				"function": "getCredentials",
				"return": "result",
				"params": {},
				"cachable": 60,
				"path": "/api/credential/get.json"
			},
			"getUserPermissions": {
				"function": "getUserPermissions",
				"return": "success",
				"params": {
					"userid": ""
				},
				"cachable": 60,
				"path": "/api/user/permissions/get.json"
			},
			"addTeslaTemplate": {
				"function": "addTeslaTemplate",
				"return": "result",
				"params": {},
				"path": "/api/tesla/transactional/add.json"
			}
		}
	}
	`

func Test_NewHTTPClient_SetsDefaults(t *testing.T) {
	url := "http://localhost:8082/"
	client := NewHTTPClient(url)

	if client.BaseURL != url {
		t.Errorf("got BaseURL %s, expected %s", client.BaseURL, url)
	}

	if client.Client == nil {
		t.Error("Client was not set")
	}
}

func Test_HTTPClient_DoFunction_QueriesFunctionsOnce(t *testing.T) {
	var functionsQueried int = 0

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		if req.URL.Path == "/api/functions.json" {
			functionsQueried++
			res.Write([]byte(functionsListJSON))
		} else if req.URL.Path == "/api/tesla/transactional/add.json" {
			assert.Equal(t, req.FormValue("name"), "hey")
			assert.Equal(t, req.FormValue("user_id"), "12")
			assert.Equal(t, req.FormValue("id"), "")

			res.Write([]byte(`{"result": "9000"}`))
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	params := url.Values{
		"user_id": {"12"},
		"name":    {"hey"},
	}
	var updatedTemplateID string
	client.DoFunction(functions()["transactional"]["addTemplate"], params, &updatedTemplateID)
	assert.Equal(t, updatedTemplateID, "9000")

	client.DoFunction(functions()["transactional"]["addTemplate"], params, &updatedTemplateID)
	assert.Equal(t, functionsQueried, 1)
}

func Test_HTTPClient_AddFunctionUpdatesTheInternalList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		assert.Equal(t, req.URL.Path, "/api/tesla/transactional/add.json")
		assert.Equal(t, req.FormValue("name"), "hey")
		assert.Equal(t, req.FormValue("user_id"), "12")
		assert.Equal(t, req.FormValue("id"), "")

		res.Write([]byte(`{"result": "9000"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	theFunc := FunctionInfo{
		Return: "result",
		Path:   "/api/tesla/transactional/add.json",
	}
	client.AddFunction(functions()["transactional"]["addTemplate"], theFunc)

	params := url.Values{
		"user_id":             {"12"},
		"name":                {"hey"},
		"default_template_id": {"template"},
	}
	var updatedTemplateID string
	client.DoFunction(functions()["transactional"]["addTemplate"], params, &updatedTemplateID)
	assert.Equal(t, updatedTemplateID, "9000")
}

func Test_HTTPClient_DoFunction_LooksUpMissingFunctions(t *testing.T) {
	var functionsQueried int = 0

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		if req.URL.Path == "/api/functions.json" {
			functionsQueried++
			res.Write([]byte(functionsListJSON))
		} else if req.URL.Path == "/api/tesla/transactional/add.json" {
			res.Write([]byte(`{"result": "9000"}`))
		} else if req.URL.Path == "/api/user/permissions/get.json" {
			res.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	theFunc := FunctionInfo{
		Return: "result",
		Path:   "/api/tesla/transactional/add.json",
	}
	client.AddFunction(functions()["transactional"]["addTemplate"], theFunc)

	var updatedTemplateID string
	client.DoFunction(functions()["transactional"]["addTemplate"], url.Values{}, &updatedTemplateID)
	assert.Equal(t, updatedTemplateID, "9000")

	assert.Equal(t, functionsQueried, 0)

	var success bool
	client.DoFunction("getUserPermissions", url.Values{}, &success)
	assert.True(t, success)

	assert.Equal(t, functionsQueried, 1)
}

type exampleJSONStructure struct {
	Foo int `json:"foo"`
}

func Test_HTTPClient_DoFunction_ReturnsErrorIfJSONMarshallingFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result": [{"foo": 1000}]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	theFunc := FunctionInfo{
		Return: "result",
		Path:   "/api/tesla/transactional/add.json",
	}
	client.AddFunction("coolFunc", theFunc)

	var structure []exampleJSONStructure
	err1 := client.DoFunction("coolFunc", url.Values{}, &structure)
	assert.NoError(t, err1)
	assert.Equal(t, structure[0].Foo, 1000)

	var wrong string
	err2 := client.DoFunction("coolFunc", url.Values{}, &wrong)
	assert.Error(t, err2)
}

type fakeHTTPRequester struct {
	callCount int
}

func (f *fakeHTTPRequester) Do(r *http.Request) (*http.Response, error) {
	f.callCount++
	return &http.Response{Body: ioutil.NopCloser(strings.NewReader(""))}, nil
}

func Test_HTTPClient_CustomClient(t *testing.T) {
	requester := &fakeHTTPRequester{}

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		if req.URL.Path == "/api/functions.json" {
			res.Write([]byte(functionsListJSON))
		} else if req.URL.Path == "/api/tesla/transactional/add.json" {
			res.Write([]byte(`{"result": "9000"}`))
		}
	}))
	defer server.Close()

	apidClient := NewHTTPClient(server.URL)
	apidClient.Client = requester

	params := url.Values{
		"user_id":             {"12"},
		"name":                {"hey"},
		"default_template_id": {"template"},
	}
	var updatedTemplateID string
	apidClient.DoFunction(functions()["transactional"]["addTemplate"], params, &updatedTemplateID)

	if requester.callCount == 0 {
		t.Error("custom client was not used")
	}
}

func Test_HTTPClient_RequestHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		if req.URL.Path == "/api/functions.json" {
			res.Write([]byte(functionsListJSON))
		} else if req.URL.Path == "/api/tesla/transactional/add.json" {
			res.Write([]byte(`{"result": "9000"}`))
		}
	}))
	defer server.Close()

	requests := make([]*http.Request, 0)

	apidClient := NewHTTPClient(server.URL)
	apidClient.RequestHandler = func(r *http.Request) {
		requests = append(requests, r)
	}

	params := url.Values{
		"user_id":             {"12"},
		"name":                {"hey"},
		"default_template_id": {"template"},
	}
	var updatedTemplateID string
	apidClient.DoFunction(functions()["transactional"]["addTemplate"], params, &updatedTemplateID)

	expectedURLs := []string{
		fmt.Sprintf("%s/api/functions.json", server.URL),
		fmt.Sprintf("%s/api/tesla/transactional/add.json?default_template_id=template&name=hey&user_id=12", server.URL),
	}

	if len(requests) != len(expectedURLs) {
		t.Errorf("got %d calls, expected %d", len(requests), len(expectedURLs))
	}

	for i, request := range requests {
		if request.Method != "GET" {
			t.Errorf("request created with wrong method. got %s, expected GET", request.Method)
		}

		if request.URL.String() != expectedURLs[i] {
			t.Errorf("got url %s, expected %s", request.URL, expectedURLs[i])
		}
	}
}

func Test_HTTPClient_Check(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		if req.URL.Path == "/api/functions.json" {
			res.Write([]byte(functionsListJSON))
		}
	}))
	defer server.Close()

	successClient := NewHTTPClient(server.URL)
	err := successClient.Check()
	if err != nil {
		t.Errorf("Check() failed: got %s, expected nil", err.Error())
	}

	failServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	}))
	defer failServer.Close()
	failClient := NewHTTPClient(failServer.URL)
	err = failClient.Check()
	if err == nil {
		t.Errorf("Check() failed: got nil error, expected an error")
	}
}
