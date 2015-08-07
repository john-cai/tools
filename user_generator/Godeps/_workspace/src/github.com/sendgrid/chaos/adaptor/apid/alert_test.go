package apidadaptor

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/stretchr/testify/assert"
)

func TestDeleteUserAlerts(t *testing.T) {
	expected := 1
	userIDs := []int{123, 234}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteUserAlerts", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.DeleteUserAlerts(userIDs)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDeleteUserAlertsError(t *testing.T) {
	expected := 0
	userIDs := []int{432, 321}
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteUserAlerts", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.DeleteUserAlerts(userIDs)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}

func TestDeleteAllUserNotificationSettings(t *testing.T) {
	expected := 1
	userIDs := []int{123, 234}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllUserNotificationSettings", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.DeleteAllUserNotificationSettings(userIDs)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDeleteAllUserNotificationSettingsError(t *testing.T) {
	expected := 0
	userIDs := []int{432, 321}
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllUserNotificationSettings", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.DeleteAllUserNotificationSettings(userIDs)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}
