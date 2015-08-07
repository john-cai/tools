package apidadaptor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/ln"
)

type UserProvisioner interface {
	UpdateUserBI(provision client.Provision) *adaptor.AdaptorError
	AddUserFingerprint(talon TalonFingerprint) (bool, *adaptor.AdaptorError)
	EditUserProfile(userProfile *client.UserProfile) (bool, *adaptor.AdaptorError)
	GenScopeSetTemplateName(userID int) (string, *adaptor.AdaptorError)
}

func (a *Adaptor) UpdateUserBI(provision client.Provision) *adaptor.AdaptorError {
	var response string
	adaptorErr := a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"user_signup_bi"},
		"where":     []string{fmt.Sprintf(`{"user_id" : "%d"}`, provision.UserProfile.UserID)},
		"values": []string{fmt.Sprintf(
			`[{"user_persona": "%s"}, {"industry": "%s"}, {"volume": "%s"}]`,
			provision.UserPersona, provision.Industry, provision.EmailVolume)}},
		&response,
	)

	if adaptorErr != nil {
		ln.Err("unable to update user signup BI", ln.Map{"err": adaptorErr.Error(), "data": provision})
		return adaptor.NewError("error storing profile information")
	}

	if response != "success" {
		ln.Err("nonsuccessful response updating user signup BI", ln.Map{"err": "unexpected result" + response, "user_id": provision.UserProfile.UserID})
		return adaptor.NewError("error storing profile information")
	}

	return nil
}

func (a *Adaptor) EditUserProfile(userProfile *client.UserProfile) (bool, *adaptor.AdaptorError) {
	var editSuccess int

	params, queryErr := query.Values(userProfile)

	if queryErr != nil {
		ln.Err("error query user profile", ln.Map{"err": queryErr.Error()})
		return false, adaptor.NewError("error query user profile")
	}

	err := a.apidClient.DoFunction("editUserProfile", params, &editSuccess)

	if err != nil {
		ln.Err("error updating user profile", ln.Map{"err": err.Error(), "user_profile": userProfile})
		return false, adaptor.NewError("error updating user profile")
	}

	if editSuccess == 0 {
		// check if the user exists
		user, adaptorError := a.GetUserProfile(userProfile.UserID)

		if adaptorError != nil {
			ln.Err("error when getting user profile", ln.Map{"err": adaptorError.Error(), "user_id": userProfile.UserID})
			return false, adaptor.NewError("error when getting user profile")
		}
		if user == nil {
			return false, adaptor.NewErrorWithStatus("user profile not found", http.StatusNotFound)
		}
		// no fields were changed
		return true, nil
	}

	return true, nil
}

func (a *Adaptor) AddUserFingerprint(talon TalonFingerprint) (bool, *adaptor.AdaptorError) {
	var success string

	v, queryErr := query.Values(talon)

	if queryErr != nil {
		ln.Err("error query talon", ln.Map{"err": queryErr.Error()})
		return false, adaptor.NewError("error query talon")
	}

	//turn timezone into string
	tz, err := json.Marshal(talon.Timezone)

	if err != nil {
		ln.Err("error when parsing talon timezone", ln.Map{"err": err.Error()})
		return false, adaptor.NewError("error parsing timezone")
	}

	v.Set("timezone", string(tz))

	columns := NewCrudColumns()
	columns.AddColumns(v)
	addErr := a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_fingerprints"},
		"values":    []string{columns.String()},
	}, &success)

	if addErr != nil {
		if strings.Contains(addErr.Error(), "key exists") {
			return true, nil
		}
		ln.Err("error adding user fingerprint", ln.Map{"err": addErr.Error()})
		formattedErr := adaptor.NewError("error adding user fingerprint")
		return false, formattedErr
	}

	if success == "success" {
		return true, nil
	}

	return false, adaptor.NewError("could not add user fingerprint")
}

type genericUserPackage struct {
	PackageID      int `json:"package_id"`
	PackageGroupID int `json:"package_group_id"`
}
type GenericAPIDResult struct {
	Name string `json:"name"`
}

func (a *Adaptor) GenScopeSetTemplateName(userID int) (string, *adaptor.AdaptorError) {
	userPkg := make([]genericUserPackage, 0)
	pkg := make([]GenericAPIDResult, 0)
	pkgGroup := make([]GenericAPIDResult, 0)

	err := a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"user_package"},
		"where":     []string{fmt.Sprintf(`{"user_id":%d}`, userID)},
	}, &userPkg)

	if err != nil || len(userPkg) != 1 {
		ln.Err("error with apid crud get for user package", ln.Map{"err": err.Error()})
		return "", adaptor.NewError("unable to get package information")
	}

	err = a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"package"},
		"where":     []string{fmt.Sprintf(`{"id":%d}`, userPkg[0].PackageID)},
	}, &pkg)

	if err != nil || len(pkg) != 1 {
		ln.Err("error with apid crud get for package", ln.Map{"err": fmt.Sprintf("%s %v", err.Error(), pkg)})
		return "", adaptor.NewError("unable to get package information")
	}

	err = a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"package_group"},
		"where":     []string{fmt.Sprintf(`{"id":%d}`, userPkg[0].PackageGroupID)},
	}, &pkgGroup)

	if err != nil || len(pkgGroup) != 1 {
		ln.Err("error with apid crud get for package group", ln.Map{"err": fmt.Sprintf("%s %v", err.Error(), pkgGroup)})
		return "", adaptor.NewError("unable to get package information")
	}

	return fmt.Sprintf("%s::%s", pkgGroup[0].Name, pkg[0].Name), nil
}
