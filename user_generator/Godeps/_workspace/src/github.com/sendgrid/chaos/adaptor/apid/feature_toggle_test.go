package apidadaptor

import (
	"errors"
	"io/ioutil"
	"net/url"
	"testing"
	"time"

	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
)

func TestFeatureEnabled(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("checkFeatureToggle", func(params url.Values) (interface{}, error) {
		return true, nil
	})

	adaptor := NewWithRetry(fakeClient, 1*time.Second, 1*time.Second)

	isEnabled, err := adaptor.IsFeatureEnabled("some_feature")

	assert.Nil(t, err, "there should be no error")
	assert.True(t, isEnabled, "feature enabled")
}

func TestFeatureDisabled(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("checkFeatureToggle", func(params url.Values) (interface{}, error) {
		return false, nil
	})

	adaptor := New(fakeClient)

	isEnabled, err := adaptor.IsFeatureEnabled("some_feature")

	assert.Nil(t, err, "there should be no error")
	assert.False(t, isEnabled, "feature disabled")
}

func TestFeatureError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("checkFeatureToggle", func(params url.Values) (interface{}, error) {
		return false, errors.New("some apid error")
	})

	adaptor := New(fakeClient)

	isEnabled, err := adaptor.IsFeatureEnabled("some_feature")

	assert.False(t, isEnabled, "feature disabled")
	assert.NotNil(t, err, "there should be an error")
}
