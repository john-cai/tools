package apidadaptor

import (
	"net/url"

	"github.com/sendgrid/chaos/adaptor"
)

type FeatureService interface {
	IsFeatureEnabled(string) (bool, *adaptor.AdaptorError)
}

func (a *Adaptor) IsFeatureEnabled(feature string) (bool, *adaptor.AdaptorError) {
	params := url.Values{
		"app_name":     []string{"chaos"},
		"feature_name": []string{feature},
	}

	var isEnabled bool

	err := a.retry("checkFeatureToggle", params, &isEnabled)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return false, formattedErr
	}

	return isEnabled, nil
}
