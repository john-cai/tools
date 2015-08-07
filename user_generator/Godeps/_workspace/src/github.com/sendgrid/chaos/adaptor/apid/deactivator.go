package apidadaptor

import "github.com/sendgrid/chaos/adaptor"

type Deactivator interface {
	GetUserPackage(userID int) (*UserPackage, *adaptor.AdaptorError)
	DeactivateUserPackage(userID int) *adaptor.AdaptorError
	InsertDeactivationReason(userID int, reason string, moving bool, inHouse bool, otherProvider string, comment string) *adaptor.AdaptorError
}
