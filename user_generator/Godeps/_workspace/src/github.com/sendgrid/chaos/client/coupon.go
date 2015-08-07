package client

type Coupon struct {
	CouponCode    string  `json:"code"`
	CouponType    string  `json:"type"`
	CouponValue   float32 `json:"value"`
	CouponPeriods int     `json:"periods"`
	Valid         bool    `json:"valid"`
}
