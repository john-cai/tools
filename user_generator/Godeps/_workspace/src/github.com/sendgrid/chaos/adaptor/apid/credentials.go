package apidadaptor

import (
	"net/url"
	"strconv"

	"github.com/sendgrid/chaos/adaptor"
)

func (a *Adaptor) GetCredentialIDs(userID int) ([]int, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
	}

	type credential struct {
		ID int `json:"id"`
	}

	var credentials []credential
	err := a.apidClient.DoFunction("getCredentials", params, &credentials)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return nil, formattedErr
	}

	ids := make([]int, len(credentials), len(credentials))
	for i, c := range credentials {
		ids[i] = c.ID
	}

	return ids, nil
}
