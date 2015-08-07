package client

type DeactivateRequestBody struct {
	Reason        string `json:"reason"`
	Password      string `json:"password"`
	Moving        bool   `json:"moving"`
	InHouse       bool   `json:"in_house"`
	OtherProvider string `json:"other_provider"`
	Comment       string `json:"comment"`
}
