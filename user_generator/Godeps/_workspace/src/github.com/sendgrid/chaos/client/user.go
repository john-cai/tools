package client

// User represents the response object for a user
type User struct {
	ID                 int    `json:"user_id,omitempty"`
	Username           string `json:"username,omitempty"`
	FirstName          string `json:"first_name,omitempty"`
	LastName           string `json:"last_name,omitempty"`
	AccountOwnerID     int    `json:"account_owner_id,omitempty"`
	AccountID          string `json:"account_id,omitempty"`
	Email              string `json:"email,omitempty"`
	Active             bool   `json:"active"`
	IsResellerDisabled int    `json:"is_reseller_disabled,omitempty"`
}

type UserProfile struct {
	UserID           int    `json:"user_id" url:"userid"`
	Phone            string `json:"phone" url:"phone,omitempty"`
	Website          string `json:"website" url:"website,omitempty"`
	FirstName        string `json:"first_name" url:"first_name,omitempty"`
	LastName         string `json:"last_name" url:"last_name,omitempty"`
	Address1         string `json:"address" url:"address,omitempty"`
	Address2         string `json:"address2" url:"address2,omitempty"`
	City             string `json:"city" url:"city,omitempty"`
	State            string `json:"state" url:"state,omitempty"`
	Zip              string `json:"zip" url:"zip,omitempty"`
	Country          string `json:"country" url:"country,omitempty"`
	Company          string `json:"company" url:"company,omitempty"`
	MultifactorPhone string `json:"multifactor_phone" url:"multifactor_phone,omitempty"`
	IsProvisionFail  int    `json:"is_provision_fail" url:"is_provision_fail,omitempty"`
}

type UserPackage struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	BasePrice          float64 `json:"base_price"`
	OveragePrice       float64 `json:"overage_price"`
	NewsletterPrice    float64 `json:"newsletter_price"`
	CampaignPrice      float64 `json:"campaign_price"`
	IsHV               bool    `json:"is_hv"`
	IsLite             bool    `json:"is_lite"`
	PackageID          string  `json:"package_id"`
	PackageStatus      string  `json:"package_status"`
	DowngradePackageID string  `json:"downgrade_package_id"`
	PlanType           string  `json:"plan_type"`
	HasIP              bool    `json:"has_ip"`
}

type Package struct {
	Credits     float64 `json:"credits"`
	Name        string  `json:"name"`
	ID          int     `json:"id"`
	HasIP       int     `json:"has_ip"`
	Permissions string  `json:"permissions"`
	IsOtis      int     `json:"is_otis"`
}

type ChangePackage struct {
	PackageID string `json:"package_id"`
	Reason    string `json:"reason"`
}

type UserResult struct {
	Result *User `json:"result"`
}

type UserProvisionStatus struct {
	Status string `json:"status"`
}
