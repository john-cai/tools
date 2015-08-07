package client

type Subuser struct {
	ID       int
	Username string
	Password string
	Email    string
	IPs      []string `json:"ips"`
}

// SubuserRequest is the request struct for getSubusers
type SubuserRequest struct {
	UserID   int
	Limit    int
	Offset   int
	Username string
}

// SubuserPermission contains the response for subuser permission check
type SubuserPermission struct {
	Message       string `json:"message"`
	HasPermission bool   `json:"has_permission"`
}
