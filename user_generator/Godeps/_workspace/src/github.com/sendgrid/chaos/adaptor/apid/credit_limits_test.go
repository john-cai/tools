package apidadaptor

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/stretchr/testify/assert"
)

func TestAddCreditLimits(t *testing.T) {
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("setUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)

	err := adaptor.AddCreditLimits(
		client.Signup{
			UserID: 180,
			FreePackage: client.PackageRecord{
				ID:             FreePackageID,
				IsFree:         1,
				PackageGroupID: FreePackageGroupID,
				Credits:        FreeAccountCreditsLimits,
			},
		})
	assert.Len(t, err, 0)
}

func TestAddCreditLimitsFailure(t *testing.T) {
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("setUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 0, errors.New("something went wrong!")
	})

	adaptor := New(fakeClient)

	err := adaptor.AddCreditLimits(
		client.Signup{
			UserID: 180,
		})
	assert.Len(t, err, 1, "apid generated an expected error")
}

func TestAddCreditLimitsFailureFreePackageInvalid(t *testing.T) {
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("setUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)

	err := adaptor.AddCreditLimits(
		client.Signup{
			UserID: 180,
		})
	assert.Len(t, err, 1, "free package doesn't have ID or Credits")
	assert.Equal(t, err[0].Error(), "free package doesn't have ID or Credits")
}

func TestDeleteCreditLimits(t *testing.T) {
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("removeUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)
	success, err := adaptor.DeleteCreditLimits(180)
	assert.Nil(t, err)
	assert.Equal(t, success, 1)
}

func TestDeleteCreditLimitsFailure(t *testing.T) {
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid error"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("removeUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 0, errors.New("apid error")
	})

	adaptor := New(fakeClient)
	var success int
	success, adaptorErr := adaptor.DeleteCreditLimits(180)
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
	assert.Equal(t, success, 0)
}

func TestSetCreditLimits(t *testing.T) {
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("setUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)
	success, err := adaptor.SetCreditLimits(180, 1000, "monthly")
	assert.Nil(t, err)
	assert.Equal(t, success, 1)
}

func TestSetCreditLimitsFailure(t *testing.T) {
	expectedErr := &adaptor.AdaptorError{
		Err:                 errors.New("apid error"),
		SuggestedStatusCode: http.StatusInternalServerError,
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("setUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 0, errors.New("apid error")
	})

	adaptor := New(fakeClient)
	var success int
	success, adaptorErr := adaptor.SetCreditLimits(180, 1000, "monthly")
	assert.Equal(t, adaptorErr.Err, expectedErr.Err)
	assert.Equal(t, adaptorErr.SuggestedStatusCode, http.StatusInternalServerError)
	assert.Equal(t, success, 0)
}
