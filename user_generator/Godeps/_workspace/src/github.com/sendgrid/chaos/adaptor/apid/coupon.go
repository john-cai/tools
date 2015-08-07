package apidadaptor

import (
	"net/url"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
)

func (a *Adaptor) CouponInfo(couponCode string) (client.Coupon, *adaptor.AdaptorError) {
	response := make([]client.Coupon, 1)
	err := a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"coupon"},
		"where":     []string{`{"code" : "` + couponCode + `"}`},
	}, &response)

	if err != nil {
		return client.Coupon{}, adaptor.NewError(err.Error())
	}

	if len(response) == 0 {
		return client.Coupon{}, adaptor.NewError("No coupons returned when searching for coupon code: " + couponCode)
	}
	couponInfo := response[0]

	return couponInfo, nil
}
