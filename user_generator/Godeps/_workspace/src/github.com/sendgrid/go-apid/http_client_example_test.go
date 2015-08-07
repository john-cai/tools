package apid_test

import (
	"net/http"
	"net/url"
	"time"

	"github.com/sendgrid/go-apid"
)

func ExampleHTTPClient() {
	// create an apid client with a custom HTTP client
	apid := apid.NewHTTPClient("http://localhost:8082")
	apid.Client = http.Client{
		Timeout: 30 * time.Second,
	}

	// add a custom header to the request
	apid.RequestHandler = func(r *http.Request) {
		r.Header.Add("Connection", "Keep-Alive")
	}

	params := url.Values{
		"userid": 180,
	}

	user := &User{}

	err := apid.DoFunction("getUserProfile", params, user)
	if err != nil {
		// do something
	}
}
