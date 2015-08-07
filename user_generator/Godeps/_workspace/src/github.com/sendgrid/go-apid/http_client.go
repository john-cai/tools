package apid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type RequestHandler func(*http.Request)

// HTTPRequester implements a Do method for retrieval of HTTP resources. See net/http's Client.
type HTTPRequester interface {
	// Requests an HTTP resource with the specified request.
	Do(*http.Request) (*http.Response, error)
}

type HTTPClient struct {
	BaseURL      string
	functions    functionsList
	hasFunctions bool

	// The client used to make the GET call. If you need certain headers or timeouts, set
	// this property.
	Client HTTPRequester

	// Allows for modification of HTTP requests before they are sent.
	RequestHandler RequestHandler
}

type FunctionInfo struct {
	Return string `json:"return"`
	Path   string `json:"path"`
}

type functionsList map[APIdFunction]FunctionInfo

func NewHTTPClient(url string) *HTTPClient {
	return &HTTPClient{BaseURL: url, functions: make(functionsList), Client: http.DefaultClient}
}

func (apid *HTTPClient) AddFunction(name APIdFunction, function FunctionInfo) {
	apid.functions[name] = function
	apid.hasFunctions = true
}

func (apid *HTTPClient) do(urlFragment string, params url.Values, dataPtr interface{}, resultName string) error {
	baseURL := fmt.Sprintf("%s%s", apid.BaseURL, urlFragment)
	return apid.makeRequestToAPId(baseURL, params, dataPtr, resultName)
}

func (apid *HTTPClient) DoFunction(name APIdFunction, params url.Values, dataPtr interface{}) error {
	thisFunc, found, err := apid.getFunctionInfo(name)

	if err != nil {
		return err
	}

	if !found {
		return errors.New(fmt.Sprintf("unknown function: %s", name))
	}

	return apid.do(thisFunc.Path, params, dataPtr, thisFunc.Return)
}

func (apid *HTTPClient) getFunctionInfo(name APIdFunction) (FunctionInfo, bool, error) {
	thisFunc, found := apid.functions[name]
	if found {
		return thisFunc, true, nil
	}

	err := apid.updateFunctionCache()
	if err != nil {
		return FunctionInfo{}, false, err
	}

	thisFunc, found = apid.functions[name]
	if found {
		return thisFunc, true, nil
	}

	return FunctionInfo{}, false, nil
}

func (apid *HTTPClient) updateFunctionCache() (err error) {
	var funcs functionsList
	err = apid.do("/api/functions.json", url.Values{}, &funcs, "functions")
	if err != nil {
		return
	}
	apid.functions = funcs
	apid.hasFunctions = true
	return
}

func (apid *HTTPClient) makeRequestToAPId(url string, params url.Values, dataPtr interface{}, resultName string) error {
	requestURL := fmt.Sprintf("%s?%s", url, params.Encode())
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return err
	}

	if apid.RequestHandler != nil {
		apid.RequestHandler(request)
	}

	response, err := apid.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var apidResponse map[string]json.RawMessage
	fullJSONUnmarshalErr := json.Unmarshal(responseBody, &apidResponse)
	if fullJSONUnmarshalErr != nil {
		return fullJSONUnmarshalErr
	}

	if response.StatusCode != http.StatusOK {
		errorString := string(apidResponse["error"])
		return errors.New(errorString)
	}

	partialJSONUnmarshalErr := json.Unmarshal(apidResponse[resultName], dataPtr)
	if partialJSONUnmarshalErr != nil {
		return partialJSONUnmarshalErr
	}

	return nil
}

/* methods below implement komodo.Healthcheck interface */
func (apid *HTTPClient) Name() string {
	return "apid"
}

func (apid *HTTPClient) Check() error {
	var funcs functionsList
	err := apid.do("/api/functions.json", url.Values{}, &funcs, "functions")
	return err
}
