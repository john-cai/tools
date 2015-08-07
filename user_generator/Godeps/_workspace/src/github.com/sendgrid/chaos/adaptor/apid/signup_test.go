package apidadaptor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expectedNewUserID := 1234

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUser", func(params url.Values) (interface{}, error) {
		return expectedNewUserID, nil
	})
	fakeClient.RegisterFunction("addUserIpGroup", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("enableUserFilter", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("addUserFilters", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("addUserPackage", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("addUserProfile", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("addBounceManagementSettings", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	fakeClient.RegisterFunction("setUserCreditLimit", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)

	createUserResult, adaptorErr := adaptor.CreateUser(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
		})

	assert.Nil(t, adaptorErr)
	assert.Equal(t, expectedNewUserID, createUserResult, "got back the expected user id")
}

func TestAddUserPartner(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})
	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// Partner credential is string
			SendGridPartner: "YTozOntzOjc6InBhcnRuZXIiO3M6MTA6InNlbmR3aXRodXMiO3M6NDoiaGFzaCI7czozOiJhYmMiO3M6MTg6InBhcnRuZXJfY3JlZGVudGlhbCI7czo4OiIxYTJiM2M0ZCI7fQ==",
		})

	assert.Nil(t, adaptorErr, fmt.Sprintf("no error should have occured, got %s", adaptorErr.Error()))

	adaptorErr = adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// Partner credential is int
			SendGridPartner: "YTozOntzOjc6InBhcnRuZXIiO3M6MTA6InNlbmR3aXRodXMiO3M6NDoiaGFzaCI7czozOiJhYmMiO3M6MTg6InBhcnRuZXJfY3JlZGVudGlhbCI7aToxMDAwMDA7fQ==",
		})

	assert.Nil(t, adaptorErr, fmt.Sprintf("no error should have occured, got %s", adaptorErr.Error()))
}

func TestAddUserPartnerWithEmptyPartnerCredential(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})
	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username:        "test_user",
			Email:           "test@example.com",
			Password:        "asdfasdf",
			SendGridPartner: "YTozOntzOjc6InBhcnRuZXIiO3M6NjoiZ29vZ2xlIjtzOjQ6Imhhc2giO047czoxODoicGFydG5lcl9jcmVkZW50aWFsIjtOO30=",
		})

	assert.Nil(t, adaptorErr, fmt.Sprintf("no error should have occured, got %s", adaptorErr.Error()))
}

func TestAddUserPartnerEmpty(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})
	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username:        "test_user",
			Email:           "test@example.com",
			Password:        "asdfasdf",
			SendGridPartner: "",
		})

	assert.NotNil(t, adaptorErr, "an error should have ocurred.")
}

func TestAddUserPartnerErr(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "i haz un error", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username:        "test_user",
			Email:           "test@example.com",
			Password:        "asdfasdf",
			SendGridPartner: "this string is not base 64 encoded",
		})
	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestAddUserPartnerErrWithJsonBody(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// bad serialized
			SendGridPartner: "e0kgYW0gbm90IGdvb2QganNvbn0=",
		})

	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestAddUserPartnerErrWithSerializeString(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// bad serialized array
			SendGridPartner: "czo0OiJhYWFhIjs=",
		})

	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestAddUserPartnerErrWithInvalidKeys(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// "partner" key is missing
			SendGridPartner: "YTozOntzOjg6InBhcnRuZXIyIjtzOjEwOiJzZW5kd2l0aHVzIjtzOjQ6Imhhc2giO3M6MzoiYWJjIjtzOjE4OiJwYXJ0bmVyX2NyZWRlbnRpYWwiO3M6ODoiMWEyYjNjNGQiO30=",
		})

	assert.NotNil(t, adaptorErr, "an error should have occured")

	adaptorErr = adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// "partner_credential" key is missing
			SendGridPartner: "YTozOntzOjc6InBhcnRuZXIiO3M6MTA6InNlbmR3aXRodXMiO3M6NDoiaGFzaCI7czozOiJhYmMiO3M6MTk6InBhcnRuZXJfY3JlZGVudGlhbDIiO3M6ODoiMWEyYjNjNGQiO30=",
		})

	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestAddUserPartnerErrWithInvalidKeyValuesType(t *testing.T) {
	expectedGetResult := make([]PartnerRecord, 1)
	expectedGetResult[0] = PartnerRecord{ID: 1, Label: "partner name"}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})
	fakeClient.RegisterFunction("get", func(param url.Values) (interface{}, error) {
		return expectedGetResult, nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// "partner" key is int
			SendGridPartner: "YTozOntzOjc6InBhcnRuZXIiO2k6MTtzOjQ6Imhhc2giO3M6MzoiYWJjIjtzOjE4OiJwYXJ0bmVyX2NyZWRlbnRpYWwiO3M6ODoiMWEyYjNjNGQiO30=",
		})

	assert.NotNil(t, adaptorErr, "an error should have occured")

	adaptorErr = adaptor.AddUserPartner(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// "partner_credential" key is array
			SendGridPartner: "YTozOntzOjc6InBhcnRuZXIiO3M6MTA6InNlbmR3aXRodXMiO3M6NDoiaGFzaCI7czozOiJhYmMiO3M6MTg6InBhcnRuZXJfY3JlZGVudGlhbCI7YToxOntpOjA7czo0OiJkYWRhIjt9fQ==",
		})

	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestUserBI(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserBI(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			SignupBI: "eyJtYyI6IkRpcmVjdCIsIm1jZCI6Imh0dHBzOi8vd3d3Lmdvb2dsZS5jb20vIiwiZ2FpZCI6bnVsbCwia3ciOm51bGx9",
		})
	assert.Nil(t, adaptorErr, "no error should have occured")
}

func TestUserBIErr(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "i haz un error", nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserBI(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			SignupBI: "this string is not base 64 encoded",
		})
	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestUserBIErrWithJsonBody(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)

	adaptorErr := adaptor.AddUserBI(
		client.Signup{
			Username: "test_user",
			Email:    "test@example.com",
			Password: "asdfasdf",
			// bad json
			SignupBI: "eyJtYyI6IkRpcmVjdCIsIm1jZCI6Imh0dHBzOi8vd3d3Lmdvb2dsZS5jb20vIg==",
		})
	assert.NotNil(t, adaptorErr, "an error should have occured")
}

func TestAddIPGroup(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	missingUserIDError := errors.New("missing userid")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUserIpGroup", func(param url.Values) (interface{}, error) {
		if param.Get("userid") == "0" {
			return 0, missingUserIDError
		}

		return 1, nil
	})

	adaptor := New(fakeClient)

	// test good case
	err := adaptor.AddIPGroup(
		client.Signup{
			UserID: 180,
		})
	assert.Len(t, err, 0)

	// test missing user error
	err = adaptor.AddIPGroup(client.Signup{})

	assert.Len(t, err, 1)
	assert.Equal(t, missingUserIDError, err[0])

}
func TestCreateUserExistsError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expectedNewUserID := 1234

	// value to be passed into apid
	expectedUsername := "test_user"

	// expected output from apidadaptor
	expectedErrMsg := "username exists"
	expectedField := "username"
	expectedStatus := http.StatusBadRequest

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUser", func(params url.Values) (interface{}, error) {
		username := params.Get("username")
		if username == expectedUsername {
			// apid returns a json hash with the key exists on duplicate entries
			return 0, fmt.Errorf("key exists")
		}
		return expectedNewUserID, nil
	})

	adaptor := New(fakeClient)

	_, adaptorErr := adaptor.CreateUser(
		client.Signup{
			Username: expectedUsername,
			Email:    "test@example.com",
			Password: "asdfasdf",
		})

	if assert.NotNil(t, adaptorErr, adaptorErr.Error()) {
		assert.Equal(t, expectedErrMsg, adaptorErr.Error())
		assert.Equal(t, expectedStatus, adaptorErr.SuggestedStatusCode)
		assert.Equal(t, expectedField, adaptorErr.Field)
	}
}

func TestCreateUserPasswordValidationError(t *testing.T) {
	expectedNewUserID := 1234

	// value to be passed into apid
	expectedUsername := "test_user"

	// expected output from apidadaptor (pass along password errors)
	expectedApidErr := "your password sucks bro. because reasons."
	expectedErr := adaptor.NewErrorWithStatus("password invalid - "+expectedApidErr, http.StatusBadRequest)

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUser", func(params url.Values) (interface{}, error) {
		username := params.Get("username")
		if username == expectedUsername {
			// apid returns a json hash with the key exists on duplicate entries
			return 0, fmt.Errorf(expectedApidErr)
		}
		return expectedNewUserID, nil
	})

	adaptor := New(fakeClient)

	_, adaptorErr := adaptor.CreateUser(
		client.Signup{
			Username: expectedUsername,
			Email:    "test@example.com",
			Password: "asdfasdf",
		})

	if assert.NotNil(t, adaptorErr, adaptorErr.Error()) {
		assert.Equal(t, expectedErr.Error(), adaptorErr.Error())
		assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorErr.SuggestedStatusCode)
	}
}

func TestCreateUserError(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expectedNewUserID := 1234

	// value to be passed into apid
	expectedUsername := "test_user"

	// expected output from apidadaptor
	expectedErr := adaptor.NewError("error when adding user")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUser", func(params url.Values) (interface{}, error) {
		username := params.Get("username")
		if username == expectedUsername {
			return 0, fmt.Errorf("some apid error")
		}
		return expectedNewUserID, nil
	})

	adaptor := New(fakeClient)

	_, adaptorErr := adaptor.CreateUser(
		client.Signup{
			Username: expectedUsername,
			Email:    "test@example.com",
			Password: "asdfasdf",
		})

	if assert.NotNil(t, adaptorErr, adaptorErr.Error()) {
		assert.Equal(t, expectedErr.Error(), adaptorErr.Error())
		assert.Equal(t, expectedErr.SuggestedStatusCode, adaptorErr.SuggestedStatusCode)
	}
}

func TestUpdateURLMailDomainSuccess(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	var actualParams url.Values
	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		actualParams = params
		return nil, nil
	})

	adaptor := New(fakeClient)
	testUserID := 1

	testUrlDomain := fmt.Sprintf("u%d.ct.sendgrid.net", testUserID)
	testMailDomain := "sendgrid.net"

	adaptor.UpdateURLMailDomain(testUserID)

	assert.Equal(t, actualParams.Get("tableName"), "user")
	assert.Equal(t, actualParams.Get("where"), fmt.Sprintf(`{"id" : "%d"}`, testUserID))
	assert.Equal(t, actualParams.Get("values"), fmt.Sprintf(`[{"url_domain": "%s"},{"mail_domain": "%s"}]`, testUrlDomain, testMailDomain))
}

func TestUpdateURLMailDomainFailure(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	var actualParams url.Values
	err := errors.New("error")

	fakeClient.RegisterFunction("update", func(params url.Values) (interface{}, error) {
		actualParams = params
		return nil, err
	})

	adaptor := New(fakeClient)
	testUserID := 1
	testUrlDomain := fmt.Sprintf("u%d.ct.sendgrid.net", testUserID)
	testMailDomain := "sendgrid.net"

	updateErr := adaptor.UpdateURLMailDomain(testUserID)

	assert.NotNil(t, updateErr)
	assert.Equal(t, actualParams.Get("tableName"), "user")
	assert.Equal(t, actualParams.Get("where"), fmt.Sprintf(`{"id" : "%d"}`, testUserID))
	assert.Equal(t, actualParams.Get("values"), fmt.Sprintf(`[{"url_domain": "%s"},{"mail_domain": "%s"}]`, testUrlDomain, testMailDomain))
}

func TestAddFiltersValid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	user := client.Signup{
		Username: "username",
		UserID:   1,
		Password: "password",
		Email:    "example@example.com",
	}

	// filter already enabled
	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("enableUserFilter", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	// no errors, but register one row updated
	fakeClient.RegisterFunction("addUserFilters", func(params url.Values) (interface{}, error) {
		return 1, nil
	})
	adaptor := New(fakeClient)
	errs := adaptor.AddFilters(user)

	assert.Len(t, errs, 0)
}

func TestAddFiltersInvalid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	user := client.Signup{
		Username: "username",
		UserID:   1,
		Password: "password",
		Email:    "example@example.com",
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("enableUserFilter", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	expectedApidError := fmt.Errorf("apid internal server error")

	fakeClient.RegisterFunction("addUserFilters", func(params url.Values) (interface{}, error) {
		return 0, expectedApidError
	})
	adaptor := New(fakeClient)
	errs := adaptor.AddFilters(user)

	assert.Len(t, errs, 4)
	assert.Equal(t, errs[0], expectedApidError)
	assert.Equal(t, errs[1], expectedApidError)
	assert.Equal(t, errs[2], expectedApidError)
	assert.Equal(t, errs[3], expectedApidError)
}

func TestAddUserProfileValid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	user := client.Signup{
		Username: "username",
		UserID:   1,
		Password: "password",
		Email:    "example@example.com",
		IP:       "192.168.1.100",
	}

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUserProfile", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)
	errs := adaptor.AddUserProfile(user)

	assert.Len(t, errs, 0)
}

func TestAddUserProfileInvalid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	user := client.Signup{
		Username: "username",
		UserID:   1,
		Password: "password",
		Email:    "example@example.com",
		IP:       "192.168.1.100",
	}

	expectedApidError := fmt.Errorf("apid internal server error")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addUserProfile", func(params url.Values) (interface{}, error) {
		return 1, expectedApidError
	})

	adaptor := New(fakeClient)
	errs := adaptor.AddUserProfile(user)

	assert.Len(t, errs, 1, "apid generated an expected error")
	assert.Equal(t, errs[0], expectedApidError)
}

func TestAddUserPackageValid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)

	err := adaptor.AddUserPackage(
		client.Signup{
			UserID:      180,
			FreePackage: client.PackageRecord{ID: 1, IsFree: 1, PackageGroupID: 1, Credits: 12000},
		})
	assert.Len(t, err, 0)
}

func TestAddUserPackageForSubuser(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	addPackageSpy := fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return "success", nil
	})

	adaptor := New(fakeClient)

	err := adaptor.AddUserPackage(
		client.Signup{
			UserID:            180,
			ResellerID:        888,
			UserPackageStatus: 444,
		})
	assert.Len(t, err, 0)

	//verify subuser params are removed
	params := addPackageSpy.CalledParams["values"][0]
	exists := strings.Contains(params, "package_id")
	assert.False(t, exists, "package id should be null")
	exists = strings.Contains(params, "package_group_id")
	assert.False(t, exists, "package group id should be null")
	exists = strings.Contains(params, "package_status")
	assert.False(t, exists, "package status should be null")

	//verify subuser params is modified
	re := regexp.MustCompile(`{"status":"(\d*)"}`)
	matches := re.FindAllStringSubmatch(params, -1)
	status := matches[0][1]
	exists = strings.Contains(params, "status")
	assert.True(t, exists, "status is passed through")
	assert.Equal(t, "444", status, "status should be pass through") //default is 1 = SendGrid Free
}

func TestAddUserPackageInvalid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	expectedApidError := fmt.Errorf("apid internal server error")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("add", func(param url.Values) (interface{}, error) {
		return 0, expectedApidError
	})

	adaptor := New(fakeClient)

	err := adaptor.AddUserPackage(
		client.Signup{
			UserID: 180,
		})
	assert.Len(t, err, 1, "apid generated an expected error")
	assert.Equal(t, err[0], expectedApidError)
}

func TestAddBounceManagementSettings(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("addBounceManagementSettings", func(params url.Values) (interface{}, error) {
		return 1, nil
	})

	adaptor := New(fakeClient)

	err := adaptor.AddBounceManagement(
		client.Signup{
			UserID: 180,
		})
	assert.Len(t, err, 0)
}

func TestGetSignupPackageInfoValid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	// Apid return free package
	fakeClient := clientfakes.NewFakeClient()

	expectedResponse := client.PackageRecord{ID: 1, IsFree: 1, PackageGroupID: 1, Credits: 12000}
	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return []client.PackageRecord{expectedResponse}, nil
	})

	adaptor := New(fakeClient)

	freePackage := adaptor.GetSignupPackageInfo("abc")
	assert.Equal(t, expectedResponse, freePackage)

	// Apid return paid package, then our function will return 12k Free package
	fakeClient = clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return []client.PackageRecord{client.PackageRecord{ID: 1, IsFree: 0, PackageGroupID: 1, Credits: 12000}}, nil
	})

	adaptor = New(fakeClient)

	freePackage = adaptor.GetSignupPackageInfo("abc")
	assert.Equal(t, client.PackageRecord{ID: FreePackageID, IsFree: 1, PackageGroupID: FreePackageGroupID, Credits: FreeAccountCreditsLimits}, freePackage)
}

func TestGetSignupPackageInfoInvalid(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()

	fakeClient.RegisterFunction("get", func(params url.Values) (interface{}, error) {
		return nil, fmt.Errorf("apid internal server error")
	})

	adaptor := New(fakeClient)

	freePackage := adaptor.GetSignupPackageInfo("abc")
	assert.Equal(t, client.PackageRecord{ID: FreePackageID, IsFree: 1, PackageGroupID: FreePackageGroupID, Credits: FreeAccountCreditsLimits}, freePackage)
}

func TestAddBounceManagementFailure(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	fakeClient := clientfakes.NewFakeClient()
	fakeClient.RegisterFunction("adBounceManagementSettings", func(params url.Values) (interface{}, error) {
		return 0, errors.New("something went wrong!")
	})

	adaptor := New(fakeClient)

	err := adaptor.AddBounceManagement(
		client.Signup{
			UserID: 180,
		})
	assert.Len(t, err, 1, "apid generated an expected error")
}
