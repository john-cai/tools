package apidadaptor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUserBISuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)

	err := adaptor.UpdateUserBI(client.Provision{})

	require.Nil(t, err, "should not err")
}

func TestUpdateUserBIWrongResult(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "some unexpected result", nil
	})

	adaptor := New(fakeClient)

	err := adaptor.UpdateUserBI(client.Provision{})

	require.NotNil(t, err, "should err")
}

func TestUpdateUserBISSomeAPIDErr(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		return "", errors.New("some apid err")
	})

	adaptor := New(fakeClient)

	err := adaptor.UpdateUserBI(client.Provision{})

	require.NotNil(t, err, "should err")
}

func TestGetTemplateNameBasedOnPackageInfo(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		if params.Get("tableName") == "user_package" {
			return []genericUserPackage{{PackageID: 653, PackageGroupID: 187}}, nil
		}
		if params.Get("tableName") == "package" {
			return []GenericAPIDResult{{Name: "Free 12k"}}, nil
		}
		if params.Get("tableName") == "package_group" {
			return []GenericAPIDResult{{Name: "Essentials/Pro"}}, nil
		}
		return "", errors.New("unexpected call to fake apid")
	})
	adaptor := New(fakeClient)

	scopeSetTemplateName, err := adaptor.GenScopeSetTemplateName(1800)

	assert.Nil(t, err, fmt.Sprintf("should not err, got `%#v`", err))
	assert.Equal(t, "Essentials/Pro::Free 12k", scopeSetTemplateName)
}

func TestAddTalonFingerprintSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "success", nil
	})
	adaptor := New(fakeClient)
	ok, adaptorErr := adaptor.AddUserFingerprint(TalonFingerprint{})

	require.Nil(t, adaptorErr, "no error should have occured")
	assert.True(t, ok, "fingerprint added")
}

func TestAddTalonFingerprintSuccessOnDuplicate(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "", errors.New("key exists")
	})
	adaptor := New(fakeClient)
	ok, adaptorErr := adaptor.AddUserFingerprint(TalonFingerprint{})

	require.Nil(t, adaptorErr, "no error should have occured")
	assert.True(t, ok, "fingerprint added")
}

func TestAddTalonFingerprintSuccessError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(params url.Values) (interface{}, error) {
		return "", errors.New("some error")
	})
	adaptor := New(fakeClient)
	ok, adaptorErr := adaptor.AddUserFingerprint(TalonFingerprint{})

	assert.NotNil(t, adaptorErr, "error should have occured")
	assert.False(t, ok, "fingerprint errored")
}
