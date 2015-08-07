package apidadaptor

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
)

type CreditService interface {
	AddCreditLimits(client.Signup) []error
	DeleteCreditLimits(int) (int, *adaptor.AdaptorError)
	GetLitePlanStartingCredits(int) (int, *adaptor.AdaptorError)
	SetCreditLimits(int, int, string) (int, *adaptor.AdaptorError)
}

func (a *Adaptor) AddCreditLimits(data client.Signup) []error {
	if data.FreePackage.ID == 0 && data.FreePackage.Credits == float64(0) {
		return []error{fmt.Errorf(`free package doesn't have ID or Credits`)}
	}

	_, err := a.SetCreditLimits(data.UserID, int(data.FreePackage.Credits), FreeAccountCreditPeriod)

	if err != nil {
		return []error{err}
	}

	return make([]error, 0)
}

func (a *Adaptor) DeleteCreditLimits(userId int) (int, *adaptor.AdaptorError) {
	var deleteResult int
	err := a.apidClient.DoFunction("removeUserCreditLimit", url.Values{
		"userid": []string{strconv.Itoa(userId)},
	}, &deleteResult)

	if err != nil {
		return 0, adaptor.NewError(err.Error())
	}

	return deleteResult, nil
}

func (a *Adaptor) GetLitePlanStartingCredits(packageID int) (int, *adaptor.AdaptorError) {
	response := make([]client.LitePlanStartingCredits, 1)
	err := a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"otis_credit_criteria"},
		"where":     []string{fmt.Sprintf(`{"criteria":"{\"age\":0}", "package_id": %d}`, packageID)},
	}, &response)

	if err != nil {
		return 0, adaptor.NewError(err.Error())
	}

	// default entry found
	if len(response) > 0 {
		return response[0].Credits, nil

		// no entry found, default to 0
	} else {
		return 0, nil
	}
}

func (a *Adaptor) SetCreditLimits(userID int, credits int, period string) (int, *adaptor.AdaptorError) {
	var success int
	err := a.apidClient.DoFunction("setUserCreditLimit", url.Values{
		"userid":     []string{strconv.Itoa(userID)},
		"credits":    []string{strconv.Itoa(credits)},
		"period":     []string{period},
		"last_reset": []string{time.Now().Format("2006-01-02")},
	}, &success)

	if err != nil {
		return 0, adaptor.NewError(err.Error())
	}
	if success == 0 {
		return 0, adaptor.NewError("could not set credit limit settings")
	}

	return success, nil
}
