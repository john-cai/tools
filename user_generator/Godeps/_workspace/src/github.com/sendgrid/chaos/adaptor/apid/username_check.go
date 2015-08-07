package apidadaptor

import (
	"fmt"
	"net/url"

	"github.com/sendgrid/chaos/adaptor"
)

var ApidErrorFmt string = "Error with apid check exists: %s"

type UsernameCheckQuery interface {
	IsUsernameAvailable(string) (bool, *adaptor.AdaptorError)
}

// IsUsernameAvailable checks if a username is available to use
func (a *Adaptor) IsUsernameAvailable(username string) (bool, *adaptor.AdaptorError) {
	var nameExists bool
	err := a.apidClient.DoFunction(
		"checkexists",
		url.Values{"username": []string{username}},
		&nameExists)

	if err != nil {
		return false, adaptor.NewError(fmt.Sprintf(ApidErrorFmt, err))
	}

	// If name exists in apid, it's not valid
	// If name doesn't exist in apid, it's valid
	return !nameExists, nil
}
