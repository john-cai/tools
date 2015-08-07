package client

const (
	ScopeMailSend                 = "mail.send"
	ScopeTriggerConfirmationEmail = "signup.trigger_confirmation"

	// Signup Pages
	ScopeProvision      = "ui.provision"
	ScopeConfirmEmail   = "ui.confirm_email"
	ScopeSignupComplete = "ui.signup_complete"

	// Profile Scopes
	ScopeProfileCreate = "user.profile.create"
	ScopeProfileRead   = "user.profile.read"
	ScopeProfileUpdate = "user.profile.update"
	ScopeProfileDelete = "user.profile.delete"

	// Profile Scopes
	ScopeEmailCreate = "user.email.create"
	ScopeEmailRead   = "user.email.read"
	ScopeEmailUpdate = "user.email.update"
	ScopeEmailDelete = "user.email.delete"

	// Package Scopes
	ScopeGetUserPackage    = "billing.read"
	ScopeUpdateUserPackage = "billing.update"
	ScopeDeleteUserPackage = "billing.delete"

	// Billing Scopes
	ScopeBillingRead   = "billing.read"
	ScopeBillingUpdate = "billing.update"

	// Subuser scopes
	ScopeSubusersCreate = "subusers.create"

	// Newsletter scopes
	ScopeNewsletterCreate = "newsletter.create"
)
