package client

type IPsResultsResponse struct {
	Result IPsResponse `json:"result"`
}

type IPsResponse struct {
	IPs []string `json:"ips"`
}
