package apidadaptor

import (
	"net/url"

	"github.com/sendgrid/chaos/adaptor"
)

type AlertService interface {
	DeleteUserAlerts([]int) (int, *adaptor.AdaptorError)
	DeleteAllUserNotificationSettings([]int) (int, *adaptor.AdaptorError)
}

func (a *Adaptor) DeleteUserAlerts(userIDs []int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userids": toStringSlice(userIDs),
	}

	var count int
	err := a.apidClient.DoFunction("deleteUserAlerts", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) DeleteAllUserNotificationSettings(userIDs []int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userids": toStringSlice(userIDs),
	}

	var count int
	err := a.apidClient.DoFunction("deleteAllUserNotificationSettings", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}
