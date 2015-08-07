package apidadaptor

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
)

func TestDeleteAllDistributorPendingUpgrades(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := 1
	userIDs := []int{123, 234}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllDistributorPendingUpgrades", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.DeleteAllDistributorPendingUpgrades(userIDs)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDeleteAllDistributorPendingUpgradesError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := 0
	userIDs := []int{432, 321}
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllDistributorPendingUpgrades", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.DeleteAllDistributorPendingUpgrades(userIDs)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}
