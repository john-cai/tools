package apid

import "net/url"

type APIdFunction string

type Client interface {
	DoFunction(name APIdFunction, params url.Values, dataPtr interface{}) error
}

// Name() and Check() methods implement komodo.Healthcheck interface (https://github.com/sendgrid/go-komodo)
type HealthcheckableClient interface {
	Client
	Name() string
	Check() error
}
