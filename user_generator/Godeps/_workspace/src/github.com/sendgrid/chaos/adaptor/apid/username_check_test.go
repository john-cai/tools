package apidadaptor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/sendgrid/go-apid/clientfakes"
	"github.com/sendgrid/ln"
	"github.com/stretchr/testify/assert"
)

var apidMockError string = "Apid Error"

func TestUsernameCheck(t *testing.T) {
	ln.SetOutput(ioutil.Discard, "test_logger")

	errorName := "errorName"
	takenName := "some_taken_username"
	availName := "some_available_username"

	takenExpected := false
	availExpected := true
	errorExpected := errors.New(fmt.Sprintf(ApidErrorFmt, apidMockError))

	fakeApid := clientfakes.NewFakeClient()
	fakeApid.RegisterFunction("checkexists", func(params url.Values) (interface{}, error) {
		username := params.Get("username")
		switch username {
		case "some_taken_username":
			return true, nil // taken
		case "some_available_username":
			return false, nil // available
		}

		return nil, errors.New(apidMockError)
	})

	adaptor := New(fakeApid)

	// Check available name
	available, adaptorErr := adaptor.IsUsernameAvailable(availName)

	assert.Nil(t, adaptorErr)
	assert.Equal(t, availExpected, available)

	// Check taken name
	available, adaptorErr = adaptor.IsUsernameAvailable(takenName)
	assert.Nil(t, adaptorErr)
	assert.Equal(t, takenExpected, available)

	// Check error case
	_, adaptorErr = adaptor.IsUsernameAvailable(errorName)
	if assert.NotNil(t, adaptorErr) {
		assert.Equal(t, errorExpected, adaptorErr.Err)
		assert.Equal(t, http.StatusInternalServerError, adaptorErr.SuggestedStatusCode)
	}
}
