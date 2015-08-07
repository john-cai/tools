package apidadaptor

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/sendgrid/chaos/adaptor"
	apid "github.com/sendgrid/go-apid"
	"github.com/sendgrid/waitfor"
)

type Adaptor struct {
	apidClient apid.Client
	interval   time.Duration
	timeout    time.Duration
}

func New(apidClient apid.Client) *Adaptor {
	return &Adaptor{apidClient: apidClient}
}

func NewWithRetry(apidClient apid.Client, interval, timeout time.Duration) *Adaptor {
	return &Adaptor{
		apidClient: apidClient,
		interval:   interval,
		timeout:    timeout,
	}
}

// retry makes a request to apid with retries on fail
func (a *Adaptor) retry(apidFunction string, params url.Values, result interface{}) *adaptor.AdaptorError {
	wrapper := func() error {
		return a.apidClient.DoFunction("checkFeatureToggle", params, result)
	}

	err := waitfor.Func(wrapper, a.interval, a.timeout)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return formattedErr
	}

	return nil
}

// Name adheres to the adminiable interface
func (a *Adaptor) Name() string {
	return "apid"
}

// Check adheres to the adminiable interface and provides the way to check that apid is working correctly
func (a *Adaptor) Check() error {
	var data interface{}
	return a.apidClient.DoFunction("getHealthcheck", nil, &data)
}

func toStringSlice(vals []int) []string {
	result := make([]string, len(vals), len(vals))
	for _, s := range vals {
		result = append(result, strconv.Itoa(s))
	}

	return result
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type CrudColumns struct {
	values []*simplejson.Json
}

func NewCrudColumns() *CrudColumns {
	return &CrudColumns{
		values: make([]*simplejson.Json, 0),
	}
}

func (c *CrudColumns) addColumn(name string, value string) {
	j := simplejson.New()
	j.Set(name, value)
	c.values = append(c.values, j)
}

func (c *CrudColumns) AddColumns(params url.Values) {
	for k, v := range params {
		c.addColumn(k, v[0])
	}
}

func (c *CrudColumns) String() string {
	d, _ := json.Marshal(c.values)

	return string(d)
}
