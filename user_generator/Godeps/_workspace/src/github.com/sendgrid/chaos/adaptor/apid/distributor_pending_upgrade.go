package apidadaptor

import (
	"net/url"

	"github.com/sendgrid/chaos/adaptor"
)

type DistributorPendingUpgradeService interface {
	DeleteAllDistributorPendingUpgrades([]int) (int, *adaptor.AdaptorError)
}

func (a *Adaptor) DeleteAllDistributorPendingUpgrades(userIDs []int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userids": toStringSlice(userIDs),
	}

	var count int
	err := a.apidClient.DoFunction("deleteAllDistributorPendingUpgrades", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}
