package apidadaptor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/stretchr/testify/assert"
)

func TestValidateIps(t *testing.T) {
	cases := []struct {
		message      string
		apidResponse int
		apidError    error
		expected     bool
		expectedErr  *adaptor.AdaptorError
		ips          []string
	}{
		{
			message:      "valid ips",
			apidResponse: 2,
			expected:     true,
			ips:          []string{"1.1.1.1", "2.2.2.2"},
		}, {
			message:     "invalid ips",
			apidError:   errors.New("ips were invalid"),
			expected:    false,
			expectedErr: adaptor.NewErrorWithStatus("One or more ips were invalid", http.StatusBadRequest),
			ips:         []string{"7.7.7.7"},
		}, {
			message:     "empty request",
			expected:    false,
			expectedErr: adaptor.NewErrorWithStatus("No ips provided", http.StatusBadRequest),
			ips:         []string{},
		}, {
			message:     "apid failure",
			apidError:   errors.New("apid failure"),
			expected:    false,
			expectedErr: adaptor.NewError("apid failure"),
			ips:         []string{"7.7.7.7"},
		}, {
			message:      "user does not have ips",
			ips:          []string{"1.1.1.1"},
			apidResponse: 0,
			expectedErr:  adaptor.NewErrorWithStatus("User does not have any IPs", http.StatusNotFound),
		},
	}

	for _, c := range cases {
		parentUserID := 180
		fakeClient := clientfakes.NewFakeClient()
		fakeClient.RegisterFunction("validateExternalIps", func(params url.Values) (interface{}, error) {
			return c.apidResponse, c.apidError
		})

		adaptor := New(fakeClient)
		actual, adaptorErr := adaptor.ValidateIPs(parentUserID, c.ips)

		// check actual response
		msg := fmt.Sprintf("validate IP response for %s", c.message)
		if !assert.Equal(t, actual, c.expected, msg) {
			return
		}

		// check error message
		if c.expectedErr != nil {
			msg = fmt.Sprintf("should have an error message for %s", c.message)
			assert.Equal(t, adaptorErr.Err, c.expectedErr.Err, msg)

			msg = fmt.Sprintf("should have an error code for %s", c.message)
			assert.Equal(t, adaptorErr.SuggestedStatusCode, c.expectedErr.SuggestedStatusCode, msg)
		}
	}
}

func TestAddUserSendIps(t *testing.T) {
	expected := 1
	userID := 123

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUserSendIp", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	// Add the IP to the user
	actual, err := adaptor.AddUserSendIP(userID, "192.168.0.1")

	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestAddUserSendIpsError(t *testing.T) {
	expected := 0
	userID := 432
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUserSendIp", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.AddUserSendIP(userID, "1.1.1.1")
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}

func TestUnassignExternalIps(t *testing.T) {
	expected := 1
	userIDs := []int{123, 234}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("unassignExternalIps", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.UnassignExternalIps(userIDs)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestUnassignExternalIpsError(t *testing.T) {
	expected := 0
	userIDs := []int{432, 321}
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("unassignExternalIps", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.UnassignExternalIps(userIDs)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}

func TestGetUserSendIps(t *testing.T) {
	expected := []string{"127.0.0.1", "192.168.0.1"}
	userID := 123

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserSendIp", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.GetUserSendIps(userID)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestGetUserSendIpsError(t *testing.T) {
	expected := []string{}
	userID := 123
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserSendIp", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.GetUserSendIps(userID)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}

func TestDeleteUserIp(t *testing.T) {
	expected := 1
	userID := 123

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteUserSendIp", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.DeleteUserIp(userID, "192.0.0.1")

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDeleteUserIpError(t *testing.T) {
	expected := 0
	userID := 123
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteUserSendIp", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.DeleteUserIp(userID, "192.0.0.1")
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}

func TestDeleteAllUserIps(t *testing.T) {
	expected := 1
	userIDs := []int{123, 234}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllUserIps", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.DeleteAllUserIps(userIDs)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDeleteAllUserIpsError(t *testing.T) {
	expected := 0
	userIDs := []int{432, 321}
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllUserIps", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.DeleteAllUserIps(userIDs)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}

func TestDeleteAllUserIpGroups(t *testing.T) {
	expected := 1
	userIDs := []int{123, 234}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllUserIpGroups", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.DeleteAllUserIpGroups(userIDs)

	assert.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDeleteAllUserIpGroupsError(t *testing.T) {
	expected := 0
	userIDs := []int{432, 321}
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid failure"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("deleteAllUserIpGroups", func(params url.Values) (interface{}, error) {
		return expected, errors.New("apid failure")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.DeleteAllUserIpGroups(userIDs)
	assert.Equal(t, actual, expected)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
}
