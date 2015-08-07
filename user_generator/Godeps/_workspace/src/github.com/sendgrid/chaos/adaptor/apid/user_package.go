package apidadaptor

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/ln"
)

const (
	DefaultSubuserLimit  = 15
	ErrorNoPackagesFound = "No packages returned when searching by uuid"
	ErrorUserNotFound    = "UserID does not exist."
)

// PackageAdjuster provides the required methods to adjust a user's package
type PackageAdjuster interface {
	GetUserPackage(userID int) (*UserPackage, *adaptor.AdaptorError)
	GetPackage(packageID int) (*Package, *adaptor.AdaptorError)
	InsertDowngradeToFreeReason(userID int, reason string) *adaptor.AdaptorError
	SetUserPackage(userID int, packageID int) *adaptor.AdaptorError
	DowngradeUserPackage(userID int, packageID int) *adaptor.AdaptorError
	PackageIDFromUUID(packageUUID string) (int, *adaptor.AdaptorError)
	GenScopeSetTemplateNameFromUserPackage(userID int, userPackage Package) (string, *adaptor.AdaptorError)
	EditUserProfile(userProfile *client.UserProfile) (bool, *adaptor.AdaptorError)
}

// UserPackageWrapper is necessary for now since the getUserPackageType endpoint
// double wraps the package info with `{"result": { "package": { ...package info ...}}}`
type UserPackageWrapper struct {
	Package *UserPackage `json:"package"`
}

// UserPackage is an internal representation of the user's package
type UserPackage struct {
	ID                      int     `json:"id"`
	UserID                  int     `json:"userid"`
	IsLite                  bool    `json:"is_lite"`
	IsHV                    bool    `json:"is_hv"`
	HasIP                   bool    `json:"has_ip"`
	SubusersLimit           int     `json:"subusers_limit"`
	Status                  int     `json:"status"`
	PackageName             string  `json:"name"`
	PackageType             string  `json:"package_type"`
	Description             string  `json:"description"`
	Price                   float64 `json:"price,string"`
	PricePerEmail           float64 `json:"price_per_email,string"`
	PricePerNewsletterEmail float64 `json:"price_per_newslettter_email,string"`
	PricePerCampaignContact float64 `json:"price_per_compaign_contact"`
	PackageUUID             string  `json:"uuid"`
	PackageStatus           string  `json:"package_status"`
	DowngradePackageUUID    string  `json:"downgrade_package_uuid"`
	Error                   string  `json:"error,omitempty"`
}

// Package represents the plan user can be subscribed to
type Package struct {
	ID           int             `json:"id"`
	GroupID      int             `json:"package_group_id"`
	Name         string          `json:"name"`
	Credits      float64         `json:"credits"`
	HasIP        int             `json:"has_ip"`
	Price        float64         `json:"price,string"`
	OveragePrice float64         `json:"price_per_email,string"`
	Permissions  map[string]bool `json:"permissions"`
	IsOtis       int             `json:"is_otis"`
}

func (a *Adaptor) SetUserPackage(userID int, packageID int) *adaptor.AdaptorError {
	var delSuccess int
	var userPackage string
	packageIDString := strconv.Itoa(packageID)

	pkg, adaptorErr := a.GetPackage(packageID)
	if adaptorErr != nil {
		return adaptorErr
	}

	// delete the user's old package
	err := a.apidClient.DoFunction("delete", url.Values{
		"tableName": []string{"user_package"},
		"where":     []string{`{"user_id" : "` + strconv.Itoa(userID) + `"}`},
	}, &delSuccess)

	if err != nil {
		ln.Err("Something went wrong trying to delete the user's current package", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError("Something went wrong trying to delete the user's current package")
	}

	params := url.Values{
		"user_id":          []string{strconv.Itoa(userID)},
		"status":           []string{strconv.Itoa(UserStatusSendGridPaid)},
		"package_id":       []string{packageIDString},
		"package_group_id": []string{strconv.Itoa(pkg.GroupID)},
		"package_status":   []string{strconv.Itoa(PackageStatusActive)},
		"start_date":       []string{time.Now().Format("2006-01-02")},
		"end_date":         []string{now.New(time.Now().AddDate(0, 1, 0)).BeginningOfMonth().Format("2006-01-02")},
		"subusers_limit":   []string{strconv.Itoa(DefaultSubuserLimit)},
		"updated_at":       []string{time.Now().String()},
	}

	columns := NewCrudColumns()
	columns.AddColumns(params)

	// add new user package
	err = a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_package"},
		"values":    []string{columns.String()},
	}, &userPackage)

	if err != nil {
		ln.Err("Something went wrong trying to add a package for the user", ln.Map{"error": err.Error(), "user_id": userID, "package_id": packageID})
		return adaptor.NewError("Something went wrong trying to add the user's package")

	}

	return nil
}

type userChurnReason struct {
	ID     int    `json:"id"`
	Reason string `json:"reason"`
}

func (a *Adaptor) getReasonID(reason string) (int, *adaptor.AdaptorError) {
	validReasons := a.GetChurnReasons()
	validReasonsList := make([]string, len(validReasons))

	for i, valid := range validReasons {
		if strings.ToLower(valid.Reason) == strings.ToLower(reason) {
			return valid.ID, nil
		}
		validReasonsList[i] = valid.Reason
	}

	msg := fmt.Sprintf("invalid reason given [%v], must be from [%s]", reason, strings.Join(validReasonsList, ", "))
	return 0, adaptor.NewErrorWithStatus(msg, http.StatusBadRequest)
}

func (a *Adaptor) GetChurnReasons() []userChurnReason {
	reasons := make([]userChurnReason, 0)
	err := a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"user_churn_reason"},
	}, &reasons)

	if err != nil {
		ln.Info("unable to get list of user_churn reasons", ln.Map{"err": err.Error()})
	}

	return reasons
}

func (a *Adaptor) InsertDowngradeToFreeReason(userID int, reason string) *adaptor.AdaptorError {
	reasonID, adaptorErr := a.getReasonID(reason)
	if adaptorErr != nil {
		return adaptorErr
	}

	params := url.Values{
		"user_id":    []string{strconv.Itoa(userID)},
		"moving":     []string{"0"},
		"event_type": []string{"Downgrade to Free"},
		"reason":     []string{strconv.Itoa(reasonID)},
	}

	columns := NewCrudColumns()
	columns.AddColumns(params)

	var ok string
	err := a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_churn"},
		"values":    []string{columns.String()},
	}, &ok)

	if err != nil {
		ln.Info("unable to call apid add on user_churn", ln.Map{"err": err.Error(), "user_id": userID})
		return adaptor.NewError("internal data storage error")
	}

	if ok != "success" {
		ln.Info("unexpected response from apid add on user_churn", ln.Map{"err": fmt.Sprintf("got '%s', want 'success'", ok), "user_id": userID})
	}

	return nil
}

func (a *Adaptor) InsertDeactivationReason(userID int, reason string, moving bool, inHouse bool, otherProvider string, comment string) *adaptor.AdaptorError {
	reasonID, adaptorErr := a.getReasonID(reason)
	if adaptorErr != nil {
		return adaptorErr
	}

	if inHouse {
		otherProvider = "in house"
	}

	competitorID, adaptorErr := a.getValidCompetitorID(otherProvider)
	if adaptorErr != nil {
		ln.Err("unable to get competitor id - "+otherProvider, ln.Map{"err": adaptorErr})
		return adaptorErr
	}
	if competitorID == 0 {
		competitorID, adaptorErr = a.insertNewCompetitor(userID, otherProvider)
		if adaptorErr != nil {
			ln.Err("unable to set new competitor id - "+otherProvider, ln.Map{"err": adaptorErr})
			return adaptorErr
		}
	}

	params := url.Values{
		"user_id":      []string{strconv.Itoa(userID)},
		"moving":       []string{strconv.Itoa(boolToInt(moving))},
		"event_type":   []string{"Cancellation"},
		"new_provider": []string{strconv.Itoa(competitorID)},
		"notes":        []string{comment},
		"reason":       []string{strconv.Itoa(reasonID)},
	}

	columns := NewCrudColumns()
	columns.AddColumns(params)

	var ok string
	err := a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_churn"},
		"values":    []string{columns.String()},
	}, &ok)

	if err != nil {
		ln.Info("unable to call apid add on user_churn", ln.Map{"err": err.Error(), "user_id": userID})
		return adaptor.NewError("internal data storage error")
	}

	if ok != "success" {
		ln.Info("unexpected response from apid add on user_churn", ln.Map{"err": fmt.Sprintf("got '%s', want 'success'", ok), "user_id": userID})
	}

	return nil
}

type competitor struct {
	ID     int    `json:"id"`
	Name   string `json:"competitor"`
	UserID int    `json:"user_id"`
}

func (a *Adaptor) getValidCompetitorID(otherProvider string) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"tableName": []string{"competitors"},
		"where":     []string{fmt.Sprintf(`{"competitor":"%s"}`, otherProvider)},
	}
	competition := []competitor{}

	err := a.apidClient.DoFunction("get", params, &competition)
	if err != nil {
		ln.Err("unable to get competitors value", ln.Map{"err": err.Error()})
		return 0, adaptor.NewError("internal data storage error")
	}

	if len(competition) == 0 {
		return 0, nil
	}

	if len(competition) >= 1 {
		// only return a valid competitor id if there is no user id
		for _, c := range competition {
			if c.UserID == 0 {
				return c.UserID, nil
			}
		}
	}
	// if we got here, it means all entries in the competitor table were tied to a user id
	// meaning the BI team has not vetted the entry
	ln.Info("no valid competitor found", nil)
	return 0, adaptor.NewError("internal data storage error")
}

func (a *Adaptor) getInvalidCompetitorID(userID int, otherProvider string) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"tableName": []string{"competitors"},
		"where":     []string{fmt.Sprintf(`{"competitor":"%s","user_id":"%d"}`, otherProvider, userID)},
	}
	competition := []competitor{}

	err := a.apidClient.DoFunction("get", params, &competition)
	if err != nil {
		ln.Err("unable to get competitors value", ln.Map{"err": err.Error()})
		return 0, adaptor.NewError("internal data storage error")
	}

	if len(competition) != 1 {
		ln.Err("did not get back single entry from competitor table", ln.Map{"err": err.Error()})
		return 0, adaptor.NewError("internal data storage error")
	}

	return competition[0].ID, nil
}

func (a *Adaptor) insertNewCompetitor(userID int, otherProvider string) (int, *adaptor.AdaptorError) {
	addParams := url.Values{
		"user_id":    []string{strconv.Itoa(userID)},
		"competitor": []string{otherProvider},
	}

	columns := NewCrudColumns()
	columns.AddColumns(addParams)

	// add new user package
	var ok string
	err := a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"competitors"},
		"values":    []string{columns.String()},
	}, &ok)

	if err != nil {
		ln.Err("unable to add to competitor table", ln.Map{"err": err.Error()})
		return 0, adaptor.NewError("internal data storage error")
	}

	if ok != "success" {
		ln.Err("error adding to competitor table", ln.Map{"err": fmt.Sprintf("%s - %s", ok, err.Error())})
		return 0, adaptor.NewError("internal data storage error")
	}

	return a.getInvalidCompetitorID(userID, otherProvider)
}

func (a *Adaptor) GetPackage(packageID int) (*Package, *adaptor.AdaptorError) {
	packageIDString := strconv.Itoa(packageID)

	var response Package
	err := a.apidClient.DoFunction("getPackage", url.Values{
		"id": []string{packageIDString},
	}, &response)

	if err != nil {
		ln.Err("Something went wrong getting a package", ln.Map{"error": err.Error(), "package_id": packageID})
		return nil, adaptor.NewError("Something went wrong getting a package")
	}

	return &response, nil
}

func (a *Adaptor) GetUserPackage(userID int) (*UserPackage, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
	}
	var userPackage UserPackageWrapper

	err := a.apidClient.DoFunction("getUserPackageType", params, &userPackage)
	if err != nil {
		ln.Err("error calling getUserPackageType", ln.Map{"error": err.Error()})
		formattedErr := adaptor.NewError(err.Error())
		return nil, formattedErr
	}

	// todo - investigate if we can remove getUserPackageType in favor of get
	//        or add package id to getUserPackageType call
	nextParams := url.Values{
		"tableName": []string{"user_package"},
		"where":     []string{fmt.Sprintf(`{"user_id":%d}`, userID)},
	}
	userPkg := []struct {
		ID int `json:"package_id"`
	}{}
	err = a.apidClient.DoFunction("get", nextParams, &userPkg)
	if err != nil {
		ln.Err("error calling get for user package", ln.Map{"error": err.Error(), "user_id": userID})
		formattedErr := adaptor.NewError("unable to process request")
		return nil, formattedErr
	}
	if len(userPkg) != 1 {
		ln.Err("did not get one result back from get call for user package", ln.Map{"user_id": userID})
		formattedErr := adaptor.NewError("unable to process request")
		return nil, formattedErr
	}

	userPackage.Package.ID = userPkg[0].ID

	return userPackage.Package, nil
}

func (a *Adaptor) PackageIDFromUUID(packageUUID string) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"tableName": []string{"package"},
		"where":     []string{fmt.Sprintf(`{"uuid":"%s"}`, packageUUID)},
	}

	var packages []struct {
		ID int `json:"id"`
	}

	err := a.apidClient.DoFunction("get", params, &packages)

	if err != nil {
		ln.Err("error calling get for user package", ln.Map{"error": err.Error(), "uuid": packageUUID})
		formattedErr := adaptor.NewError("unable to process request")
		return -1, formattedErr

	}

	if len(packages) != 1 {
		ln.Err("did not get one result back from get call for user package", ln.Map{"uuid": packageUUID})
		formattedErr := adaptor.NewError(ErrorNoPackagesFound)
		return -1, formattedErr
	}

	return packages[0].ID, nil
}

func (a *Adaptor) DowngradeUserPackage(userID int, packageID int) *adaptor.AdaptorError {
	var result string
	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"user_package"},
		"where":     []string{fmt.Sprintf(`{"user_id":"%d"}`, userID)},
		"values":    []string{fmt.Sprintf(`[{"upgrade_package_id":"%d"}, {"package_status":"%d"}]`, packageID, PackageStatusPendingDowngrade)},
	}, &result)
	if err != nil {
		ln.Err("Error downgrading user package", ln.Map{"error": err.Error(), "user_id": userID, "package_id": packageID})
		return adaptor.NewError(err.Error())
	}
	return nil
}

func (a *Adaptor) DeactivateUserPackage(userID int) *adaptor.AdaptorError {
	var result string
	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"user_package"},
		"where":     []string{fmt.Sprintf(`{"user_id":"%d"}`, userID)},
		"values":    []string{fmt.Sprintf(`[{"package_status":"%d"}]`, PackageStatusPendingCancellation)},
	}, &result)
	if err != nil {
		ln.Err("Error deactivating user package", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError(err.Error())
	}
	return nil
}

func (a *Adaptor) GenScopeSetTemplateNameFromUserPackage(userID int, userPackage Package) (string, *adaptor.AdaptorError) {
	pkgGroup := make([]GenericAPIDResult, 0)

	err := a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"package_group"},
		"where":     []string{fmt.Sprintf(`{"id":%d}`, userPackage.GroupID)},
	}, &pkgGroup)

	if err != nil || len(pkgGroup) != 1 {
		ln.Err("error with apid crud get for package group", ln.Map{"error": fmt.Sprintf("%s %v", err.Error(), pkgGroup)})
		return "", adaptor.NewError("unable to get package information")
	}

	return fmt.Sprintf("%s::%s", pkgGroup[0].Name, userPackage.Name), nil
}
