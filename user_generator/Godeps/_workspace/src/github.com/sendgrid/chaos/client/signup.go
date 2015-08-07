package client

// CreditAllocationType is an enumeration of the different
// credit allocation types for a subuser.
type CreditAllocationType string

const (
	CreditAllocationTypeUnlimited CreditAllocationType = "unlimited"
)

// PackageRecord represents the package which will be assign on signup
type PackageRecord struct {
	ID             int     `json:"id"`
	IsFree         int     `json:"is_free"`
	PackageGroupID int     `json:"package_group_id"`
	Credits        float64 `json:"credits"`
}

// Signup represents the payload to create a user in the adaptor
type Signup struct {
	Username          string
	UserID            int
	Password          string
	Email             string
	IP                string
	ResellerID        int
	OutboundClusterID int
	UserPackageStatus int
	Active            bool

	// package record
	FreePackage PackageRecord

	// optional parameter for business intelligence
	// base64 encoded json from tracking cookie
	SignupBI string

	// optional parameter for partner tracking
	// base64 encoded json from sendgrid_partner cookie
	SendGridPartner string
}

// SignupResponse contains the very valuable information
type SignupResponse struct {
	Username         string            `json:"username"`
	UserID           int               `json:"user_id"`
	Email            string            `json:"email"`
	SGToken          string            `json:"signup_session_token"`
	Token            string            `json:"authorization_token"`
	CreditAllocation *CreditAllocation `json:"credit_allocation,omitempty"`
}

type BillingCallbackResponseWrapper struct {
	Result SuccessResponse `json:"result"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type CreditAllocation struct {
	Type CreditAllocationType `json:"type"`
}

type AccountConfirmation struct {
	SGToken         string `json:"signup_session_token"`
	PackageUUID     string `json:"package_id"`
	PaymentMethodID string `json:"payment_method_id"`
	CouponCode      string `json:"coupon_code"`
}

type AccountSubscription struct {
	SGToken         string `json:"signup_session_token"`
	PackageID       int    `json:"package_id"`
	PaymentMethodID string `json:"payment_method_id"`
	Coupon          Coupon `json:"coupon,omitempty"`
}

type SendConfirmation struct {
	Email string `json:"email"`
}

type Provision struct {
	UserProfile UserProfile `json:"user_profile,omitempty"`
	Talon       string      `json:"talon,omitempty"`
	EmailVolume string      `json:"email_volume,omitempty"`
	Industry    string      `json:"industry,omitempty"`
	UserPersona string      `json:"occupation,omitempty"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
