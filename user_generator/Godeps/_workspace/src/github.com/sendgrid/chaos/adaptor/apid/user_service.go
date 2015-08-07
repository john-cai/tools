package apidadaptor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/go-apid"
	"github.com/sendgrid/ln"
)

// user represents the APID response object
type User struct {
	ID                 int    `json:"id"`
	Username           string `json:"username"`
	AccountOwnerID     int    `json:"reseller_id"`
	AccountID          string `json:"account_id"`
	Email              string `json:"email"`
	Active             int    `json:"active"`
	IsResellerDisabled int    `json:"is_reseller_disabled"`
}

func (u *User) isActive() bool {
	if u.Active == 1 {
		return true
	}
	return false
}

type TalonFingerprint struct {
	UserID      int                      `json:"user_id" url:"user_id"`
	Version     int                      `json:"version" url:"version"`
	Timezone    TalonFingerprintTimezone `json:"tz" url:"timezone"`
	Language    string                   `json:"lang" url:"language"`
	UserAgent   string                   `json:"ua" url:"user_agent"`
	Fingerprint string                   `json:"fp" url:"fingerprint"`
}

type TalonFingerprintTimezone struct {
	DST  bool `json:"dst"`
	TZO  int  `json:"tzo"`
	STZO int  `json:"stzo"`
}

type UserHolds map[string]interface{}

type UserService interface {
	ActivateUser(int) *adaptor.AdaptorError
	EditUser(*client.User) (bool, *adaptor.AdaptorError)
	GetUser(int) (*client.User, *adaptor.AdaptorError)
	GetUserByUsername(string) (*client.User, *adaptor.AdaptorError)
	GetUserPackage(int) (*UserPackage, *adaptor.AdaptorError)
	GetUserProfile(int) (*client.UserProfile, *adaptor.AdaptorError)
	SoftDeleteUser(int) (int, *adaptor.AdaptorError)
	WebsiteActivateUser(int) *adaptor.AdaptorError
}

type ProvisionStatusService interface {
	GetUserHolds(int) (UserHolds, *adaptor.AdaptorError)
}

// GetUser populates the response struct User for use in the chaos application
func (a *Adaptor) GetUser(userID int) (*client.User, *adaptor.AdaptorError) {
	functions := []apid.APIdFunction{"getUserInfo", "getUserInfoMaster"}
	for _, fn := range functions {
		var apidUserResults User
		err := a.apidClient.DoFunction(fn, url.Values{"userid": []string{strconv.Itoa(userID)}}, &apidUserResults)

		if err != nil {
			ln.Err(fmt.Sprintf("unable to %s", string(fn)), ln.Map{"err": err.Error(), "user_id": userID})
			formattedErr := adaptor.NewError(err.Error())

			if strings.Contains(err.Error(), "Invalid user id:") {
				formattedErr = userNotFoundError(userID)
			}
			return nil, formattedErr
		}

		if apidUserResults.ID != 0 {
			user := client.User{
				ID:                 apidUserResults.ID,
				Username:           apidUserResults.Username,
				AccountOwnerID:     apidUserResults.AccountOwnerID,
				AccountID:          apidUserResults.AccountID,
				Email:              apidUserResults.Email,
				Active:             apidUserResults.isActive(),
				IsResellerDisabled: apidUserResults.IsResellerDisabled,
			}
			return &user, nil
		}
	}

	return nil, userNotFoundError(userID)
}

// SoftDeleteUser performs a soft delete on a given user ID.
func (a *Adaptor) SoftDeleteUser(userID int) (int, *adaptor.AdaptorError) {
	var apidUserResults int
	err := a.apidClient.DoFunction("softDeleteUser", url.Values{"userid": []string{strconv.Itoa(userID)}}, &apidUserResults)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	if apidUserResults == 0 {
		return 0, userNotFoundError(userID)
	}

	return apidUserResults, nil
}

// EditUser will update user data in the database
// Note that this will not return an error if the user doesn't exist
func (a *Adaptor) EditUser(user *client.User) (bool, *adaptor.AdaptorError) {
	var apidUserResults int

	params := url.Values{
		"userid": []string{strconv.Itoa(user.ID)},
	}

	if user.AccountOwnerID != 0 {
		params.Add("reseller_id", strconv.Itoa(user.AccountOwnerID))
	}
	if user.Username != "" {
		params.Add("username", user.Username)
	}
	if user.Email != "" {
		params.Add("email", user.Email)
	}

	err := a.apidClient.DoFunction("editUser", params, &apidUserResults)

	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return false, formattedErr
	}

	// If apidUserResults == 0, then no updates were made to the db
	// so either the user doesn't exist or we are inserting the same
	// username, email, reseller_id that are already in the db

	return true, nil
}

func (a *Adaptor) GetUserProfile(userID int) (*client.UserProfile, *adaptor.AdaptorError) {
	var userProfile client.UserProfile

	err := a.apidClient.DoFunction("getUserProfile", url.Values{"userid": []string{strconv.Itoa(userID)}}, &userProfile)
	if err != nil {
		ln.Err("error getting user profile", ln.Map{"error": err.Error(), "user_id": userID})
		return nil, adaptor.NewError("error getting user profile")
	}

	return &userProfile, nil
}

// GetUserByUsername populates the response struct User for use in the chaos application
func (a *Adaptor) GetUserByUsername(username string) (*client.User, *adaptor.AdaptorError) {
	functions := []apid.APIdFunction{"getUserInfo", "getUserInfoMaster"}
	for _, fn := range functions {
		var apidUserResults User
		err := a.apidClient.DoFunction(fn, url.Values{"username": []string{username}}, &apidUserResults)

		if err != nil {
			formattedErr := adaptor.NewError(err.Error())
			return nil, formattedErr
		}

		if apidUserResults.ID != 0 {
			user := client.User{
				ID:             apidUserResults.ID,
				Username:       apidUserResults.Username,
				AccountOwnerID: apidUserResults.AccountOwnerID,
				AccountID:      apidUserResults.AccountID,
				Email:          apidUserResults.Email,
			}
			return &user, nil
		}
	}

	msg := fmt.Sprintf("the resource does not exist for username: %s", username)
	return nil, adaptor.NewErrorWithStatus(msg, http.StatusNotFound)
}

// ActivateUser sets the user_profile.activated = 1
func (a *Adaptor) ActivateUser(userID int) *adaptor.AdaptorError {
	var ignored interface{}

	whereMap := map[string]interface{}{
		"user_id": strconv.Itoa(userID),
	}

	where, _ := json.Marshal(whereMap)
	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": {"user_profile"},
		"values":    {`[{"activated":"1"},{"website_activated":"1"}]`},
		"where":     []string{string(where)},
	}, &ignored)

	if err != nil {
		return adaptor.NewError(err.Error())
	}

	return nil
}

func (a *Adaptor) WebsiteActivateUser(userID int) *adaptor.AdaptorError {
	var ignored interface{}

	whereMap := map[string]interface{}{
		"user_id": strconv.Itoa(userID),
	}

	where, _ := json.Marshal(whereMap)
	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": {"user_profile"},
		"values":    {`[{"website_activated":"1"}]`},
		"where":     []string{string(where)},
	}, &ignored)

	if err != nil {
		return adaptor.NewError(err.Error())
	}

	return nil
}

func (a *Adaptor) GetUserHolds(userID int) (UserHolds, *adaptor.AdaptorError) {
	userHolds := make(UserHolds)
	err := a.apidClient.DoFunction("getUserHolds", url.Values{"userid": []string{strconv.Itoa(userID)}}, &userHolds)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return nil, formattedErr
	}

	return userHolds, nil
}

func userNotFoundError(userID int) *adaptor.AdaptorError {
	msg := fmt.Sprintf("the resource does not exist for id: %d", userID)
	return adaptor.NewErrorWithStatus(msg, http.StatusNotFound)
}
