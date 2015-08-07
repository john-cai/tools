package apidadaptor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23

	expected := User{ID: userID, Username: "mrpickles", AccountOwnerID: 4000, AccountID: "iamanaccountid"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.GetUser(userID)

	require.Nil(t, adaptorErr)
	require.NotNil(t, actual)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.AccountOwnerID, actual.AccountOwnerID)
	assert.Equal(t, expected.AccountID, actual.AccountID)
}

func TestGetUserProfile(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23

	expected := client.UserProfile{
		UserID:          userID,
		Phone:           "123-123-1234",
		Website:         "www.google.com",
		FirstName:       "Homer",
		LastName:        "Simpson",
		Address1:        "12345 Seasame St.",
		City:            "Springfield",
		State:           "IL",
		Zip:             "11252",
		Country:         "USA",
		Company:         "Simpsons",
		IsProvisionFail: 0,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserProfile", func(params url.Values) (interface{}, error) {
		userIDFromParams, _ := strconv.Atoi(params.Get("userid"))

		switch userIDFromParams {
		case userID:
			return expected, nil
		}
		return nil, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.GetUserProfile(userID)

	require.Nil(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, actual, &expected)

}

func TestGetUserProfileUserNotFound(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	ln.SetOutput(ioutil.Discard, "test_logger")

	someOtherUserID := 24

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserProfile", func(params url.Values) (interface{}, error) {
		return nil, errors.New("some serious error happened")
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.GetUserProfile(someOtherUserID)

	require.NotNil(t, err)
	require.Nil(t, actual)

}

func TestEditUserProfile(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := client.UserProfile{
		UserID:    23,
		Phone:     "123-123-1234",
		Website:   "www.google.com",
		FirstName: "Homer",
		LastName:  "Simpson",
		Address1:  "12345 Seasame St.",
		City:      "Springfield",
		State:     "IL",
		Zip:       "11252",
		Country:   "USA",
		Company:   "Simpsons",
	}
	toModify := client.UserProfile{}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUserProfile", func(params url.Values) (interface{}, error) {
		userID, _ := strconv.Atoi(params.Get("user_id"))
		toModify.UserID = userID
		toModify.Phone = params.Get("phone")
		toModify.Website = params.Get("website")
		toModify.FirstName = params.Get("first_name")
		toModify.LastName = params.Get("last_name")
		toModify.Address1 = params.Get("address")
		toModify.City = params.Get("city")
		toModify.State = params.Get("state")
		toModify.Zip = params.Get("zip")
		toModify.Country = params.Get("country")
		toModify.Company = params.Get("company")
		return 1, nil
	})

	adaptor := New(fakeClient)

	success, err := adaptor.EditUserProfile(&expected)

	require.Nil(t, err)
	assert.True(t, success)
	assert.Equal(t, expected.Phone, toModify.Phone)
	assert.Equal(t, expected.Website, toModify.Website)
	assert.Equal(t, expected.FirstName, toModify.FirstName)
	assert.Equal(t, expected.LastName, toModify.LastName)
	assert.Equal(t, expected.Address1, toModify.Address1)
	assert.Equal(t, expected.City, toModify.City)
	assert.Equal(t, expected.State, toModify.State)
	assert.Equal(t, expected.Zip, toModify.Zip)
	assert.Equal(t, expected.Country, toModify.Country)
	assert.Equal(t, expected.Company, toModify.Company)
}
func TestEditUserProfileError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	ln.SetOutput(ioutil.Discard, "test_logger")

	someProfile := client.UserProfile{
		UserID:    23,
		Phone:     "123-123-1234",
		Website:   "www.google.com",
		FirstName: "Homer",
		LastName:  "Simpson",
		Address1:  "12345 Seasame St.",
		City:      "Springfield",
		State:     "IL",
		Zip:       "11252",
		Country:   "USA",
		Company:   "Simpsons",
	}
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUserProfile", func(params url.Values) (interface{}, error) {
		return 0, errors.New("some serious error happened")
	})

	adaptor := New(fakeClient)

	success, err := adaptor.EditUserProfile(&someProfile)

	require.NotNil(t, err)
	assert.False(t, success)
}

func TestGetEmptyUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23

	expected := User{ID: userID}

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		userIDFromParams, _ := strconv.Atoi(params.Get("userid"))
		switch userIDFromParams {
		case userID:
			return expected, nil
		}

		return nil, fmt.Errorf("not a valid call")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.GetUser(userID)

	require.Nil(t, adaptorErr)
	require.NotNil(t, actual)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.AccountOwnerID, actual.AccountOwnerID)
	assert.Equal(t, expected.AccountID, actual.AccountID)
}

func TestGetUserSlaveReadFailMasterReadSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23

	bad := User{ID: 0}
	expected := User{ID: userID, Username: "mrpickles", AccountOwnerID: 4000, AccountID: "iamanaccountid"}

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		return bad, nil
	})

	fakeClient.RegisterFunction("getUserInfoMaster", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.GetUser(userID)

	require.Nil(t, adaptorErr)
	require.NotNil(t, actual)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.AccountOwnerID, actual.AccountOwnerID)
	assert.Equal(t, expected.AccountID, actual.AccountID)
}

func TestGetUserUnknownError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expectedErr := adaptor.NewError("apid fails")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		return nil, errors.New("apid fails")
	})

	adaptor := New(fakeClient)

	_, adaptorErr := adaptor.GetUser(123)

	assert.Equal(t, expectedErr.Error(), adaptorErr.Error())
	assert.Equal(t, http.StatusInternalServerError, adaptorErr.SuggestedStatusCode)
}

func TestAddUserFingerprint(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	tz := TalonFingerprintTimezone{
		DST:  true,
		TZO:  240,
		STZO: 300,
	}
	talonFingerprint := TalonFingerprint{
		UserID:      23,
		Version:     2,
		Timezone:    tz,
		Language:    "en-US",
		UserAgent:   "Mozilla/5.0",
		Fingerprint: "fingerprint",
	}

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		columns := params.Get("values")

		if strings.Contains(columns, `{"language":"en-US"}`) &&
			strings.Contains(columns, `{"fingerprint":"fingerprint"}`) {
			return "success", nil
		}

		return false, errors.New("invalid values")
	})

	adaptor := New(fakeClient)

	success, err := adaptor.AddUserFingerprint(talonFingerprint)

	require.Nil(t, err)
	require.True(t, success)

}

func TestSoftDeleteUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23

	expected := 1

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("softDeleteUser", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, adaptorErr := adaptor.SoftDeleteUser(userID)

	assert.NoError(t, adaptorErr)
	assert.Equal(t, expected, actual)
}

func TestSoftDeleteUserUnknownUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	expectedErr := errors.New("the resource does not exist for id: 123")

	fakeClient.RegisterFunction("softDeleteUser", func(params url.Values) (interface{}, error) {
		return 0, nil
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.SoftDeleteUser(123)

	assert.Equal(t, expectedErr.Error(), adaptorErr.Error())
	assert.Equal(t, http.StatusNotFound, adaptorErr.SuggestedStatusCode)
	assert.Equal(t, 0, actual)
}

func TestSoftDeleteUserError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23
	expectedErr := adaptor.NewError("apid fails")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("softDeleteUser", func(params url.Values) (interface{}, error) {
		return 0, errors.New("apid fails")
	})

	adaptor := New(fakeClient)

	actual, adaptorErr := adaptor.SoftDeleteUser(userID)

	assert.Equal(t, expectedErr.Error(), adaptorErr.Error())
	assert.Equal(t, http.StatusInternalServerError, adaptorErr.SuggestedStatusCode)
	assert.Equal(t, 0, actual)
}

func TestEditUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23
	var actualParams url.Values

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUser", func(params url.Values) (interface{}, error) {
		actualParams = params
		return 1, nil
	})

	adaptor := New(fakeClient)
	request := &client.User{
		ID:             userID,
		AccountOwnerID: 123,
		Username:       "jake_the_dog",
		Email:          "jake@dog.com",
	}

	actual, adaptorErr := adaptor.EditUser(request)

	assert.NoError(t, adaptorErr)
	assert.Equal(t, actualParams, url.Values{"userid": []string{strconv.Itoa(userID)}, "reseller_id": []string{"123"}, "username": []string{"jake_the_dog"}, "email": []string{"jake@dog.com"}})
	assert.True(t, actual)
}

func TestEditUserNoUpdate(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	// When a call is made to update a user, but there are no changes in the db
	// (If we are editing to the same info the user already has in the db)
	userID := 23
	var actualParams url.Values

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("editUser", func(params url.Values) (interface{}, error) {
		actualParams = params
		return 0, nil
	})

	adaptor := New(fakeClient)

	request := &client.User{
		ID:             userID,
		AccountOwnerID: 123,
		Username:       "jake_the_dog",
		Email:          "jake@dog.com",
	}

	actual, adaptorErr := adaptor.EditUser(request)

	assert.NoError(t, adaptorErr)
	assert.Equal(t, actualParams, url.Values{"userid": []string{strconv.Itoa(userID)}, "reseller_id": []string{"123"}, "username": []string{"jake_the_dog"}, "email": []string{"jake@dog.com"}})
	assert.True(t, actual)
}

func TestEditUserError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	userID := 23
	expectedErr := adaptor.NewError("apid fails")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("editUser", func(params url.Values) (interface{}, error) {
		return 0, errors.New("apid fails")
	})

	adaptor := New(fakeClient)
	actual, adaptorErr := adaptor.EditUser(&client.User{ID: userID})

	assert.Equal(t, expectedErr.Error(), adaptorErr.Error())
	assert.Equal(t, http.StatusInternalServerError, adaptorErr.SuggestedStatusCode)
	assert.False(t, actual)
}

func TestGetUserByUsername(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	username := "mrpickles"

	expected := User{ID: 123, Username: username, AccountOwnerID: 4000, AccountID: "iamanaccountid"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, clientErr := adaptor.GetUserByUsername(username)

	assert.Nil(t, clientErr)
	require.NotNil(t, actual)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.AccountOwnerID, actual.AccountOwnerID)
	assert.Equal(t, expected.AccountID, actual.AccountID)
}

func TestGetUserByUsernameUnknownError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errorMessage := errors.New("failure to apid")
	expectedErr := &adaptor.AdaptorError{Err: errorMessage, SuggestedStatusCode: http.StatusInternalServerError}

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		return nil, errorMessage
	})

	adaptor := New(fakeClient)
	_, clientErr := adaptor.GetUserByUsername("mrpickles")

	assert.Equal(t, expectedErr, clientErr)
	assert.Equal(t, http.StatusInternalServerError, clientErr.SuggestedStatusCode)
}

func TestGetUserByUsernameSlaveFailMasterSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	username := "mrpickles"

	bad := User{ID: 0}
	expected := User{ID: 123, Username: username, AccountOwnerID: 4000, AccountID: "iamanaccountid"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserInfo", func(params url.Values) (interface{}, error) {
		return bad, nil
	})

	fakeClient.RegisterFunction("getUserInfoMaster", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	actual, clientErr := adaptor.GetUserByUsername(username)

	assert.Nil(t, clientErr)
	require.NotNil(t, actual)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.AccountOwnerID, actual.AccountOwnerID)
	assert.Equal(t, expected.AccountID, actual.AccountID)
}

func TestActivateUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := "success"

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return expected, nil
	})

	adaptor := New(fakeClient)
	clientErr := adaptor.ActivateUser(180)

	assert.Nil(t, clientErr)
}

func TestActivateUserError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errorMessage := errors.New("failure to apid")
	expectedErr := &adaptor.AdaptorError{Err: errorMessage, SuggestedStatusCode: http.StatusInternalServerError}

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return nil, errorMessage
	})

	adaptor := New(fakeClient)
	clientErr := adaptor.ActivateUser(180)

	assert.Equal(t, expectedErr, clientErr)
	assert.Equal(t, http.StatusInternalServerError, clientErr.SuggestedStatusCode)
}

func TestGetUserHolds(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	var userIDFromParams int

	expected := UserHolds{
		"10":  "test",
		"101": "test2",
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserHolds", func(params url.Values) (interface{}, error) {
		userIDFromParams, _ = strconv.Atoi(params.Get("userid"))
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.GetUserHolds(23)

	require.Nil(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, actual, expected)
	assert.Equal(t, userIDFromParams, 23)
}

func TestGetUserHoldsEmpty(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	var userIDFromParams int
	expected := make(UserHolds)

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserHolds", func(params url.Values) (interface{}, error) {
		userIDFromParams, _ = strconv.Atoi(params.Get("userid"))
		return expected, nil
	})

	adaptor := New(fakeClient)

	actual, err := adaptor.GetUserHolds(24)

	require.Nil(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, actual, expected)
	assert.Equal(t, userIDFromParams, 24)
}
