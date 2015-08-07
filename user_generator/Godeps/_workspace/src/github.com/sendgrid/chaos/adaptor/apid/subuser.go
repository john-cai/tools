package apidadaptor

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
)

type SubuserService interface {
	CountSubusers(*client.SubuserRequest) (int, *adaptor.AdaptorError)
	GetSubuserIDs(int) ([]int, *adaptor.AdaptorError)
	SoftDeleteSubusers(int) (int, *adaptor.AdaptorError)
}

type EnableService interface {
	Enable(userID int, enable bool) (bool, *adaptor.AdaptorError)
}

func (a *Adaptor) GetSubusers(data *client.SubuserRequest) ([]client.Subuser, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(data.UserID)},
	}
	if data.Limit > 0 {
		params.Set("limit", strconv.Itoa(data.Limit))
	}
	if data.Offset > 0 {
		params.Set("offset", strconv.Itoa(data.Offset))
	}
	if data.Username != "" {
		params.Set("username", data.Username)
	}

	var subusers []client.Subuser
	err := a.apidClient.DoFunction("getSubusers", params, &subusers)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return nil, formattedErr
	}

	return subusers, nil
}

func (a *Adaptor) CountSubusers(data *client.SubuserRequest) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(data.UserID)},
	}

	if data.Username != "" {
		params.Set("username", data.Username)
	}

	var count int
	err := a.apidClient.DoFunction("countSubusers", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) SoftDeleteSubusers(parentID int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(parentID)},
	}

	var count int
	err := a.apidClient.DoFunction("softDeleteSubusers", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) GetSubuserIDs(parentID int) ([]int, *adaptor.AdaptorError) {
	params := url.Values{
		"reseller_id": []string{strconv.Itoa(parentID)},
	}

	var ids []int
	err := a.apidClient.DoFunction("getUseridsByReseller", params, &ids)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return nil, formattedErr
	}

	return ids, nil
}

func (a *Adaptor) Enable(userID int, enable bool) (bool, *adaptor.AdaptorError) {
	params := url.Values{
		"userid":               []string{strconv.Itoa(userID)},
		"active":               {fmt.Sprintf("%d", boolToInt(enable))},
		"is_reseller_disabled": {fmt.Sprintf("%d", boolToInt(!enable))},
	}

	var success int
	err := a.apidClient.DoFunction("editUser", params, &success)
	if err != nil {
		formattedErr := &adaptor.AdaptorError{
			Err:                 err,
			SuggestedStatusCode: http.StatusInternalServerError,
		}
		return false, formattedErr
	}

	return success != 0, nil
}
