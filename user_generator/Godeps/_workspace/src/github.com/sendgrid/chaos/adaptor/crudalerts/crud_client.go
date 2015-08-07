package crudalerts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/ln"
)

const (
	defaultUserThreshold = 90
)

var CRUDAlertsError = errors.New("could not reach crud alerts")

type CrudClient interface {
	SetUsageNotifications(signup client.Signup) []error
	Check() error
	Name() string
}

type crudClient struct {
	crudAlertHost            string
	crudAlertPort            int
	crudAlertHealthcheckPort int
}

type AlertData struct {
	Type       string `json:"type"`
	EmailTo    string `json:"email_to"`
	Percentage int    `json:"percentage"`
}

func NewCrudClient(host string, port int, healthcheckPort int) *crudClient {
	return &crudClient{
		crudAlertHost:            host,
		crudAlertPort:            port,
		crudAlertHealthcheckPort: healthcheckPort,
	}
}
func (c *crudClient) Name() string {
	return "crud_alerts"
}
func (c *crudClient) Check() error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/healthcheck", c.crudAlertHost, c.crudAlertHealthcheckPort))

	if err != nil || resp.StatusCode != http.StatusOK {
		ln.Err(
			"error reaching crud_alerts",
			ln.Map{
				"error":                       err.Error(),
				"crud_alert_host":             c.crudAlertHost,
				"crud_alert_healthcheck_port": c.crudAlertHealthcheckPort,
			},
		)
		return CRUDAlertsError
	}
	defer resp.Body.Close()

	return nil
}
func (c *crudClient) SetUsageNotifications(signup client.Signup) []error {
	data, _ := json.Marshal(
		AlertData{
			Type:       "usage_limit",
			EmailTo:    signup.Email,
			Percentage: defaultUserThreshold,
		})

	url := fmt.Sprintf("http://%s:%d/users/%d/alerts", c.crudAlertHost, c.crudAlertPort, signup.UserID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return []error{fmt.Errorf("error when setting usage notifications %s", err.Error())}
	}
	defer resp.Body.Close()
	return []error{}

}

// for acceptance test purposes
func (c *crudClient) GetUsageNotifications(signup client.SignupResponse) ([]AlertData, error) {
	var alerts []AlertData
	url := fmt.Sprintf("http://%s:%d/users/%d/alerts", c.crudAlertHost, c.crudAlertPort, signup.UserID)

	resp, err := http.Get(url)
	if err != nil {
		return []AlertData{}, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&alerts)

	if err != nil {
		return []AlertData{}, err
	}
	return alerts, nil
}
