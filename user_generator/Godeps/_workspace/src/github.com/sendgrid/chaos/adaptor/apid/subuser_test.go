package apidadaptor

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSubusers(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := []client.Subuser{
		{
			ID:       1,
			Username: "dustin",
			Email:    "dustin@dustin.com",
		}, {
			ID:       2,
			Username: "jovel",
			Email:    "jovel@jovel.com",
		},
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getSubusers", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.GetSubusers(&client.SubuserRequest{UserID: 180})

	assert.Nil(t, adaptorError)
	if assert.NotNil(t, actual) {
		assert.Equal(t, len(expected), len(actual))
		for i, _ := range actual {
			assert.Equal(t, expected[i].ID, actual[i].ID)
			assert.Equal(t, expected[i].Username, actual[i].Username)
			assert.Equal(t, expected[i].Email, actual[i].Email)
		}
	}
}

func TestGetSubusersEmpty(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := []client.Subuser{}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getSubusers", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.GetSubusers(&client.SubuserRequest{UserID: 180})

	assert.Nil(t, adaptorError)
	assert.Equal(t, len(expected), len(actual))
}

func TestGetSubusersError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errMsg := errors.New("apid fails")
	expectedErr := &adaptor.AdaptorError{Err: errMsg, SuggestedStatusCode: http.StatusInternalServerError}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getSubusers", func(params url.Values) (interface{}, error) {
		return nil, errMsg
	})

	adaptor := New(fakeClient)
	_, adaptorError := adaptor.GetSubusers(&client.SubuserRequest{UserID: 180})

	assert.Equal(t, expectedErr.Err, adaptorError.Err)
	assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorError.SuggestedStatusCode)
}

func TestCountSubusers(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := 2

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("countSubusers", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.CountSubusers(&client.SubuserRequest{UserID: 180})

	assert.Nil(t, adaptorError)
	assert.Equal(t, expected, actual)
}

func TestCountSubusersError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errMsg := errors.New("apid fails")
	expectedErr := adaptor.NewError("apid fails")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("countSubusers", func(params url.Values) (interface{}, error) {
		return 0, errMsg
	})

	adaptor := New(fakeClient)
	_, adaptorError := adaptor.CountSubusers(&client.SubuserRequest{UserID: 180})

	assert.Equal(t, expectedErr.Err, adaptorError.Err)
	assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorError.SuggestedStatusCode)
}

func TestSoftDeletSubusers(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := 2

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("softDeleteSubusers", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.SoftDeleteSubusers(123)

	assert.NoError(t, adaptorError)
	assert.Equal(t, expected, actual)
}

func TestSoftDeleteSubusersError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errMsg := errors.New("apid fails")
	expectedErr := adaptor.NewError("apid fails")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("softDeleteSubusers", func(params url.Values) (interface{}, error) {
		return 0, errMsg
	})

	adaptor := New(fakeClient)
	_, adaptorError := adaptor.SoftDeleteSubusers(123)

	assert.Equal(t, expectedErr.Err, adaptorError.Err)
	assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorError.SuggestedStatusCode)
}

func TestEnable(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := 1
	var apidParams url.Values

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUser", func(params url.Values) (interface{}, error) {
		apidParams = params
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.Enable(123, true)

	require.Nil(t, adaptorError)
	require.True(t, actual)
	require.Equal(t, apidParams["userid"], []string{"123"})
	require.Equal(t, apidParams["active"], []string{"1"})
	require.Equal(t, apidParams["is_reseller_disabled"], []string{"0"})
}

func TestEnable_Disable(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := 1
	var apidParams url.Values

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUser", func(params url.Values) (interface{}, error) {
		apidParams = params
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.Enable(123, false)

	require.Nil(t, adaptorError)
	require.True(t, actual)
	require.Equal(t, apidParams["userid"], []string{"123"})
	require.Equal(t, apidParams["active"], []string{"0"})
	require.Equal(t, apidParams["is_reseller_disabled"], []string{"1"})
}

func TestEnableApidError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errMsg := errors.New("apid fail")
	expectedError := &adaptor.AdaptorError{Err: errMsg, SuggestedStatusCode: http.StatusInternalServerError}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUser", func(params url.Values) (interface{}, error) {
		return 0, errMsg
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.Enable(123, true)

	require.Equal(t, expectedError, adaptorError)
	require.False(t, actual)
}

func TestGetSubuserIDs(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := []int{1, 2}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUseridsByReseller", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.GetSubuserIDs(180)

	assert.Nil(t, adaptorError)
	require.NotNil(t, actual)
	assert.Equal(t, []int{1, 2}, actual)
}

func TestGetSubuserIDsEmpty(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := []int{}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUseridsByReseller", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorError := adaptor.GetSubuserIDs(180)

	assert.Nil(t, adaptorError)
	assert.Equal(t, []int{}, actual)
}

func TestGetSubuserIDsError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errMsg := errors.New("apid fails")
	expectedErr := &adaptor.AdaptorError{Err: errMsg, SuggestedStatusCode: http.StatusInternalServerError}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUseridsByReseller", func(params url.Values) (interface{}, error) {
		return nil, errMsg
	})

	adaptor := New(fakeClient)
	_, adaptorError := adaptor.GetSubuserIDs(180)

	assert.Equal(t, expectedErr.Err, adaptorError.Err)
	assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorError.SuggestedStatusCode)
}
