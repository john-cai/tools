package apidadaptor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserPackage(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expected := UserPackageWrapper{
		Package: &UserPackage{
			UserID:        180,
			IsLite:        false,
			IsHV:          true,
			SubusersLimit: 5,
			Status:        3,
		},
	}
	expectedPackageCall := []struct {
		ID int `json:"package_id"`
	}{
		{ID: 2},
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserPackageType", func(params url.Values) (interface{}, error) {
		return expected, nil
	})
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return expectedPackageCall, nil
	})

	adaptor := New(fakeClient)

	parentID := 180
	actual, adaptorErr := adaptor.GetUserPackage(parentID)

	assert.Nil(t, adaptorErr)
	if assert.NotNil(t, actual) {
		assert.Equal(t, expected.Package.UserID, actual.UserID)
		assert.Equal(t, expected.Package.SubusersLimit, actual.SubusersLimit)
		assert.Equal(t, expected.Package.IsLite, actual.IsLite)
		assert.Equal(t, expected.Package.Status, actual.Status)
		assert.Equal(t, expected.Package.IsHV, actual.IsHV)
	}
}

func TestGetUserPackageError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errMsg := "apid fails"
	parentID := 180
	expectedErr := adaptor.NewError(errMsg)

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("getUserPackageType", func(params url.Values) (interface{}, error) {
		return nil, errors.New(errMsg)
	})

	adaptor := New(fakeClient)

	_, adaptorError := adaptor.GetUserPackage(parentID)

	assert.Equal(t, expectedErr.Err, adaptorError.Err)
	assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorError.SuggestedStatusCode)
}

func TestSetUserPackageSuccess(t *testing.T) {
	userID := 180
	packageID := 11
	packageGroupID := 2

	addMethodCalled := false

	fakeClient := clientfakes.NewFakeClient()

	expectedPackage := Package{ID: packageID, GroupID: packageGroupID}
	fakeClient.RegisterFunction("getPackage", func(params url.Values) (interface{}, error) {
		return expectedPackage, nil
	})

	fakeClient.RegisterFunction("delete", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		addMethodCalled = true
		assert.True(t, strings.Contains(params.Get("values"), strconv.Itoa(packageGroupID)), "the package group uuid was looked up and stored")
		return "success", nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.SetUserPackage(userID, packageID)

	require.Nil(t, adaptorErr, "expected no error, got "+adaptorErr.Error())
	assert.True(t, addMethodCalled, "expect add method called in apid because tests are run in that fake")
}

func TestSetUserPackageError(t *testing.T) {
	userID := 180
	packageID := 11
	packageGroupID := 2

	fakeClient := clientfakes.NewFakeClient()

	expectedPackage := Package{ID: packageID, GroupID: packageGroupID}
	fakeClient.RegisterFunction("getPackage", func(params url.Values) (interface{}, error) {
		return expectedPackage, nil
	})

	fakeClient.RegisterFunction("delete", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "", errors.New("some apid error with add to user package")
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.SetUserPackage(userID, packageID)

	require.NotNil(t, adaptorErr, "expected error")
}

func TestInsertDowngradeToFreeReasonsSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	expectedValidReasons := []userChurnReason{
		userChurnReason{ID: 1, Reason: "reason_a"},
		userChurnReason{ID: 2, Reason: "reason_b"},
	}
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return expectedValidReasons, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)
	reason := "reason_a"
	adaptorErr := adaptor.InsertDowngradeToFreeReason(1800, reason)

	require.Nil(t, adaptorErr, fmt.Sprintf("expected no error, got %#v", adaptorErr))
}

func TestInsertDowngradeToFreeReasonInvalidReason(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	expectedValidReasons := []userChurnReason{
		userChurnReason{ID: 1, Reason: "reason_a"},
		userChurnReason{ID: 2, Reason: "reason_b"},
	}
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return expectedValidReasons, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)
	reason := "reason_c"
	adaptorErr := adaptor.InsertDowngradeToFreeReason(1800, reason)

	require.NotNil(t, adaptorErr, "should error")
}

func TestInsertDowngradeToFreeReasonApidErr(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	expectedValidReasons := []userChurnReason{
		userChurnReason{ID: 1, Reason: "reason_a"},
		userChurnReason{ID: 2, Reason: "reason_b"},
	}
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return expectedValidReasons, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "", errors.New("some apid error adding to user_churn")
	})

	adaptor := New(fakeClient)
	reason := "reason_c"
	adaptorErr := adaptor.InsertDowngradeToFreeReason(1800, reason)

	require.NotNil(t, adaptorErr, "should error")
}

func TestInsertDeactivationReaonsSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	addMethodCalled := false
	firstCompetitorGetCall := true

	fakeClient := clientfakes.NewFakeClient()
	expectedValidReasons := []userChurnReason{
		userChurnReason{ID: 1, Reason: "reason_a"},
		userChurnReason{ID: 2, Reason: "reason_b"},
	}

	// in addition to doing a successful deactivation result,
	// we are also testing that we can insert a new record into
	// the competitor table.
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "competitors" {
			if firstCompetitorGetCall {
				firstCompetitorGetCall = false
				return []competitor{}, nil
			}
			// any subsequent call to get competitors
			return []competitor{competitor{ID: 3}}, nil
		}
		if params.Get("tableName") == "user_churn_reason" {
			return expectedValidReasons, nil
		}

		return "", errors.New("unexpected table called")
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "competitors" {
			addMethodCalled = true
			t.Log("this should have the new competitor in it:", params.Get("values"))
			assert.True(t, strings.Contains(params.Get("values"), "Some New Competitor"))
			return "success", nil
		}
		return "success", nil
	})

	adaptor := New(fakeClient)
	reason := "reason_a"
	moving := true
	inHouse := false
	adaptorErr := adaptor.InsertDeactivationReason(1800, reason, moving, inHouse, "Some New Competitor", "notes on why")

	assert.True(t, addMethodCalled)
	require.Nil(t, adaptorErr, fmt.Sprintf("expected no error, got %#v", adaptorErr.Error()))
}

func TestInsertDeactivationReaonsInvalidReason(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	expectedValidReasons := []userChurnReason{
		userChurnReason{ID: 1, Reason: "reason_a"},
		userChurnReason{ID: 2, Reason: "reason_b"},
	}
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "competitors" {
			return []competitor{competitor{ID: 3}}, nil
		}
		return expectedValidReasons, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "competitors" {
			return "success", nil
		}
		return "success", nil
	})

	adaptor := New(fakeClient)
	reason := "reason_c"
	moving := true
	inHouse := true
	adaptorErr := adaptor.InsertDeactivationReason(1800, reason, moving, inHouse, "new provider", "notes on why")

	require.NotNil(t, adaptorErr, "should error")
}

func TestInsertDeactivationReaonsApidErr(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	expectedValidReasons := []userChurnReason{
		userChurnReason{ID: 1, Reason: "reason_a"},
		userChurnReason{ID: 2, Reason: "reason_b"},
	}
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "competitors" {
			return []competitor{competitor{ID: 3}}, nil
		}
		return expectedValidReasons, nil
	})
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "competitors" {
			return "success", nil
		}
		return "", errors.New("some apid error adding to user_churn")
	})

	adaptor := New(fakeClient)
	reason := "reason_a"
	moving := true
	inHouse := true
	adaptorErr := adaptor.InsertDeactivationReason(1800, reason, moving, inHouse, "new provider", "notes on why")

	require.NotNil(t, adaptorErr, "should error")
}

func TestDowngradeUserPackageSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)
	adaptorErr := adaptor.DowngradeUserPackage(1800, 11)

	require.Nil(t, adaptorErr, fmt.Sprintf("got %#v, but should not error", adaptorErr.Error()))
}

func TestDowngradeUserPackageError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "", errors.New("some apid error for downgrade user package")
	})

	adaptor := New(fakeClient)
	adaptorErr := adaptor.DowngradeUserPackage(1800, 11)

	require.NotNil(t, adaptorErr, "should error")
}

func TestDeactivateUserPackageSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)
	adaptorErr := adaptor.DeactivateUserPackage(1800)

	require.Nil(t, adaptorErr, fmt.Sprintf("got %#v, but should not error", adaptorErr.Error()))
}

func TestDeactivateUserPackageError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "", errors.New("some apid error for downgrade user package")
	})

	adaptor := New(fakeClient)
	adaptorErr := adaptor.DeactivateUserPackage(1800)

	require.NotNil(t, adaptorErr, "should error")
}
