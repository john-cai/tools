package apidadaptor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/now"
	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/ln"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
)

const (
	// filter constants
	ClickTrackFilterID   = 4
	ClickTrackFilterName = "clicktrack"
	OpenTrackFilterID    = 6
	OpenTrackFilterName  = "opentrack"
	DKIMFilterID         = 21
	DKIMFilterName       = "dkim"
	DomainKeysFilterID   = 12
	DomainKeysFilterName = "domainkeys"

	FreeAccountCreditsLimits = 12000
	FreeAccountCreditPeriod  = "monthly"
	LitePlanCreditPeriod     = "daily"
	DKIMDomain               = "sendgrid.me"

	UserUrlDomain  = "ct.sendgrid.net"
	UserMailDomain = "sendgrid.net"

	// user status codes
	UserStatusSendGridUnlimited   = 0
	UserStatusSendGridFree        = 1
	UserStatusSendGridPayAsYouGo  = 2
	UserStatusSendGridPaid        = 3
	UserStatusSendGridPaidSubUser = 4
	UserStatusReseller            = 5
	UserStatusResellerPaid        = 6
	UserStatusResellerFree        = 7
	UserStatusResellerPaidSubUser = 8

	FreePackageID      = 653
	FreePackageGroupID = 186

	PackageStatusInactive                 = 0
	PackageStatusActive                   = 1
	PackageStatusDisabled                 = 2
	PackageStatusUnpaid                   = 3
	PackageStatusPendingUpgrade           = 4
	PackageStatusPendingCancellation      = 5
	PackageStatusPendingDowngrade         = 6
	PackageStatusResellerAccountUpgrade   = 7 // legacy term: distributor
	PackageStatusResellerAccountDowngrade = 8 // legacy term: distributor
	PackageStatusResellerAccountClose     = 9 // legacy term: distributor

	ApidSuccess = "success"

	ErrorDataStorage = "there has been an error with data storage"
)

type Filter struct {
	Id       int
	Name     string
	Settings string
	IsValid  bool
}

type UserCreator interface {
	CreateUser(client.Signup) (int, *adaptor.AdaptorError)
	AddFilters(client.Signup) []error
	AddIPGroup(client.Signup) []error
	AddUserPackage(client.Signup) []error
	AddUserProfile(client.Signup) []error
	AddBounceManagement(client.Signup) []error
	AddCreditLimits(client.Signup) []error
	AddUserBI(client.Signup) *adaptor.AdaptorError
	UpdateURLMailDomain(int) *adaptor.AdaptorError
	AddUserPartner(client.Signup) *adaptor.AdaptorError
	GetSignupPackageInfo(string) client.PackageRecord
}

func (a *Adaptor) CreateUser(data client.Signup) (int, *adaptor.AdaptorError) {
	today := time.Now().Format("2006-01-02")
	params := url.Values{
		"username":             []string{data.Username},
		"password":             []string{data.Password},
		"email":                []string{data.Email},
		"reputation_start":     []string{today},
		"last_reputation_warn": []string{today},
	}

	if data.Active {
		params.Add("active", "1")
	}

	if data.ResellerID != 0 {
		params.Set("reseller_id", strconv.Itoa(data.ResellerID))
		params.Set("outbound_cluster_id", strconv.Itoa(data.OutboundClusterID))
	}

	var userID int
	err := a.apidClient.DoFunction("addUser", params, &userID)
	if err != nil {
		formattedErr := adaptor.NewError("error when adding user")
		ln.Err("error when adding user", ln.Map{"error": err.Error()})
		if strings.Contains(err.Error(), "key exists") {
			formattedErr = &adaptor.AdaptorError{
				Err:                 errors.New("username exists"),
				Field:               "username",
				SuggestedStatusCode: http.StatusBadRequest,
			}
		}
		if strings.Contains(err.Error(), "password") {
			// assume the error is a password validation error
			errStr := strings.ToLower(strings.Replace(err.Error(), `"`, ``, -1))
			errStr = fmt.Sprintf("password invalid - %s", errStr)
			formattedErr = &adaptor.AdaptorError{
				Err:                 errors.New(errStr),
				SuggestedStatusCode: http.StatusBadRequest,
				Field:               "password",
			}
		}
		return 0, formattedErr
	}

	data.UserID = userID
	return userID, nil
}

func (a *Adaptor) UpdateURLMailDomain(userID int) *adaptor.AdaptorError {
	var response string

	url_domain := fmt.Sprintf("u%d.%s", userID, UserUrlDomain)

	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"user"},
		"where":     []string{fmt.Sprintf(`{"id" : "%d"}`, userID)},
		"values":    []string{fmt.Sprintf(`[{"url_domain": "%s"},{"mail_domain": "%s"}]`, url_domain, UserMailDomain)}}, &response)

	if err != nil {
		ln.Err("error when updating url and mail domain", ln.Map{"error": err.Error()})
		return adaptor.NewError("error when updating url and mail domain")
	}

	return nil
}

func (a *Adaptor) AddFilters(data client.Signup) []error {
	var wg sync.WaitGroup

	filters := make([]*Filter, 0)
	filters = append(filters, &Filter{
		Id:       DKIMFilterID,
		Name:     DKIMFilterName,
		Settings: fmt.Sprintf(`{"domain": "%s", "use_from": 0}`, DKIMDomain),
		IsValid:  true,
	})
	filters = append(filters, &Filter{
		Id:       DomainKeysFilterID,
		Name:     DomainKeysFilterName,
		Settings: fmt.Sprintf(`{"domain": "%s", "use_from":0, "sender": 1}`, DKIMDomain),
		IsValid:  true,
	})
	filters = append(filters, &Filter{
		Id:       ClickTrackFilterID,
		Name:     ClickTrackFilterName,
		Settings: fmt.Sprintf(`{"enable_text": 1}`),
		IsValid:  true,
	})
	filters = append(filters, &Filter{
		Id:       OpenTrackFilterID,
		Name:     OpenTrackFilterName,
		Settings: "",
		IsValid:  true,
	})

	errs := make([]error, 0)
	errChan := make(chan error)

	for _, filterEntry := range filters {
		wg.Add(1)

		go func(userId int, f *Filter) {
			err := a.createFilter(userId, f)

			if err != nil {
				errChan <- err

				return
			}
			wg.Done()

		}(data.UserID, filterEntry)
	}

	go func() {
		for {
			err := <-errChan

			if err != nil {
				errs = append(errs, err)
			}
			wg.Done()
		}
	}()

	wg.Wait()

	return errs
}

func (a *Adaptor) AddUserPackage(data client.Signup) []error {

	params := url.Values{
		"user_id":          []string{strconv.Itoa(data.UserID)},
		"status":           []string{strconv.Itoa(UserStatusSendGridFree)},
		"package_id":       []string{strconv.Itoa(data.FreePackage.ID)},
		"package_group_id": []string{strconv.Itoa(data.FreePackage.PackageGroupID)},
		"package_status":   []string{strconv.Itoa(PackageStatusActive)},
		"start_date":       []string{time.Now().Format("2006-01-02")},
		"end_date":         []string{now.New(time.Now().AddDate(0, 1, 0)).BeginningOfMonth().Format("2006-01-02")},
		"subusers_limit":   []string{"0"},
		"updated_at":       []string{time.Now().String()},
	}

	if data.ResellerID != 0 {
		params.Set("status", strconv.Itoa(data.UserPackageStatus))
		params.Del("package_id")
		params.Del("package_group_id")
		params.Del("package_status")
	}

	columns := NewCrudColumns()
	columns.AddColumns(params)

	var success string
	err := a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_package"},
		"values":    []string{columns.String()},
	}, &success)

	if err != nil {
		return []error{err}
	}

	if success != ApidSuccess {
		return []error{errors.New("add user package unsuccessful")}
	}

	return make([]error, 0)
}

func (a *Adaptor) AddIPGroup(data client.Signup) []error {
	var success int
	err := a.apidClient.DoFunction("addUserIpGroup", url.Values{
		"userid":  []string{strconv.Itoa(data.UserID)},
		"groupid": []string{"1"},
	}, &success)

	if err != nil {
		return []error{err}
	}

	return make([]error, 0)
}

func (a *Adaptor) AddUserProfile(data client.Signup) []error {
	var success int
	err := a.apidClient.DoFunction("addUserProfile", url.Values{
		"userid":            []string{strconv.Itoa(data.UserID)},
		"registration_ip":   []string{data.IP},
		"activated":         []string{"0"},
		"website_activated": []string{"0"},
	}, &success)

	if err != nil {
		return []error{err}
	}
	if success == 0 {
		return []error{fmt.Errorf("could not insert into user profile table")}
	}

	return make([]error, 0)
}

func (a *Adaptor) AddBounceManagement(data client.Signup) []error {
	var success int
	err := a.apidClient.DoFunction("addBounceManagementSettings", url.Values{
		"userid":        []string{strconv.Itoa(data.UserID)},
		"block_expired": []string{"true"},
	}, &success)

	if err != nil {
		return []error{err}
	}
	if success == 0 {
		return []error{fmt.Errorf("could not add bounce settings")}
	}

	return make([]error, 0)
}

func (a *Adaptor) createFilter(userId int, f *Filter) error {
	var row_count int

	// enable filter
	err := a.apidClient.DoFunction("enableUserFilter", url.Values{
		"userid": []string{strconv.Itoa(userId)},
		"type":   []string{f.Name},
	}, &row_count)

	if err != nil {
		// only skip error if filter already enabled
		if strings.Contains(err.Error(), "already enabled for user") {
			ln.Debug("user filter already enabled", ln.Map{"debug": err})
		} else {
			return err
		}
	}

	// create filters settings
	err = a.apidClient.DoFunction("addUserFilters", url.Values{
		"userid":   []string{strconv.Itoa(userId)},
		"filterid": []string{strconv.Itoa(f.Id)},
		"settings": []string{f.Settings},
	}, &row_count)

	return err
}

type PartnerRecord struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

func (a *Adaptor) AddUserPartner(signup client.Signup) *adaptor.AdaptorError {
	data, err := base64.StdEncoding.DecodeString(padBase64(signup.SendGridPartner))
	if err != nil {
		ln.Err("unable to decode sendgrid partner data", ln.Map{"err": err.Error(), "value": signup.SendGridPartner})
		return adaptor.NewError(ErrorDataStorage)
	}

	decoder := php_serialize.NewUnSerializer(string(data))
	decodedSendGridPartner, err := decoder.Decode()
	if err != nil {
		ln.Err(fmt.Sprintf("unable to unserialize data for signup sendgridPartner - '%s'", string(data)), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	// Deserialize php object
	var partnerName string
	var partnerCredential string

	sendGridPartnerPhpArray, ok := decodedSendGridPartner.(php_serialize.PhpArray)

	if !ok {
		errMsg := fmt.Sprintf("Unable to convert %v to PhpArray", decodedSendGridPartner)
		ln.Err(errMsg, ln.Map{"err": errMsg})
		return adaptor.NewError(ErrorDataStorage)
	}

	if partnerNameObject, ok := sendGridPartnerPhpArray["partner"]; !ok {
		errMsg := "Array value decoded incorrectly, key `partner` doest not exists"
		ln.Err(errMsg, ln.Map{"err": errMsg})
		return adaptor.NewError(ErrorDataStorage)
	} else if partnerName, ok = partnerNameObject.(string); !ok {
		errMsg := fmt.Sprintf("Unable to convert %v to string", partnerNameObject)
		ln.Err(errMsg, ln.Map{"err": errMsg})
		return adaptor.NewError(ErrorDataStorage)
	}

	if partnerCredentialObject, ok := sendGridPartnerPhpArray["partner_credential"]; !ok {
		errMsg := "Array value decoded incorrectly, key `partner_credential` doest not exists"
		ln.Err(errMsg, ln.Map{"err": errMsg})
		return adaptor.NewError(ErrorDataStorage)
	} else if partnerCredentialObject == nil {
		partnerCredential = ""
	} else if partnerCredentialInt, ok := partnerCredentialObject.(int); !ok {
		if partnerCredential, ok = partnerCredentialObject.(string); !ok {
			errMsg := fmt.Sprintf("Unable to convert %v to string or int", partnerCredentialObject)
			ln.Err(errMsg, ln.Map{"err": errMsg})
			return adaptor.NewError(ErrorDataStorage)
		}
	} else {
		partnerCredential = strconv.Itoa(partnerCredentialInt)
	}

	getResult := make([]PartnerRecord, 0)

	err = a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"partner"},
		"where":     []string{fmt.Sprintf(`{"label":"%s"}`, partnerName)},
	}, &getResult)
	if err != nil {
		ln.Err("unable to fetch partner details for label", ln.Map{"err": err.Error(), "partner": partnerName})
		return adaptor.NewError(ErrorDataStorage)
	}

	if len(getResult) != 1 {
		ln.Err(fmt.Sprintf("wrong number or results for partner search. got %d, want 1", len(getResult)), ln.Map{"err": err.Error(), "partner": partnerName})
		return adaptor.NewError(ErrorDataStorage)
	}

	type UserPartner struct {
		UserID            int    `json:"user_id"`
		PartnerID         int    `json:"partner_id"`
		PartnerCredential string `json:"partner_credential"`
	}
	insertData := UserPartner{
		UserID:            signup.UserID,
		PartnerID:         getResult[0].ID,
		PartnerCredential: partnerCredential,
	}

	jsonBytes, err := json.Marshal([]UserPartner{insertData})

	if err != nil {
		ln.Err(fmt.Sprintf("unable to marshal insertData"), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	// store the partner data
	var addResult string
	err = a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_partner"},
		"values":    []string{string(jsonBytes)},
	}, &addResult)

	if err != nil {
		ln.Err(fmt.Sprintf("unable to call apid add method on user_partner"), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	if addResult != ApidSuccess {
		ln.Err(fmt.Sprintf("apid set for user_partner returned '%s', want 'success'", addResult), nil)
		return adaptor.NewError(ErrorDataStorage)
	}

	return nil
}

// AddUserBI inspects the signup bi data that is base64 encoded
// and decodes the json therein and stores it in the db
func (a *Adaptor) AddUserBI(signup client.Signup) *adaptor.AdaptorError {
	// marketingInfo represents the json that will be base64ecoded as SignupBI
	// when creating a user
	type marketingInfo struct {
		MarketingChannel       string `json:"mc"`
		MarketingChannelDetail string `json:"mcd"`
	}

	// formattedBI has the proper name mapping for the db
	type formattedBI struct {
		UserID                 int    `json:"user_id,omitempty"`
		LeadSource             string `json:"lead_source,omitempty"`
		MarketingChannel       string `json:"marketing_channel,omitempty"`
		MarketingChannelDetail string `json:"marketing_channel_detail,omitempty"`
		EmailVolume            string `json:"volume,omitempty"`
		Industry               string `json:"industry,omitempty"`
		Occupation             string `json:"user_persona,omitempty"`
		InitialPackage         int    `json:"initial_package,omitempty"`
	}

	if signup.SignupBI == "" {
		ln.Err("attempted to decode empty bi data", nil)
		return adaptor.NewError(ErrorDataStorage)
	}

	data, err := base64.StdEncoding.DecodeString(padBase64(signup.SignupBI))
	if err != nil {
		ln.Err("unable to decode signup bi", ln.Map{"err": err.Error(), "value": signup.SignupBI})
		return adaptor.NewError(ErrorDataStorage)
	}

	BI := marketingInfo{}
	err = json.Unmarshal(data, &BI)
	if err != nil {
		ln.Err("unable to unmarshal data for signup bi", ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	insertData := formattedBI{
		UserID:                 signup.UserID,
		LeadSource:             "Free Account Signup",
		MarketingChannel:       BI.MarketingChannel,
		MarketingChannelDetail: BI.MarketingChannelDetail,
		InitialPackage:         signup.FreePackage.ID,
	}

	jsonBytes, err := json.Marshal([]formattedBI{insertData})

	if err != nil {
		ln.Err(fmt.Sprintf("unable to marshal insertData"), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	// store the BI data
	var addResult string
	err = a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"user_signup_bi"},
		"values":    []string{string(jsonBytes)},
	}, &addResult)

	if err != nil {
		ln.Err(fmt.Sprintf("unable to call apid add method on user_signup_bi"), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	if addResult != ApidSuccess {
		ln.Err(fmt.Sprintf("apid set for user_signup_bi returned '%s', want 'success'", addResult), nil)
		return adaptor.NewError(ErrorDataStorage)
	}

	return nil
}

func (a *Adaptor) UpdateUserInitialPackageBI(userID int, packageID int) *adaptor.AdaptorError {
	// formattedBI has the proper name mapping for the db
	type formattedBI struct {
		InitialPackage int `json:"initial_package,omitempty"`
	}

	updateData := formattedBI{
		InitialPackage: packageID,
	}

	jsonBytes, err := json.Marshal([]formattedBI{updateData})

	if err != nil {
		ln.Err(fmt.Sprintf("unable to marshal updateData"), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	// store the BI data
	var addResult string
	err = a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"user_signup_bi"},
		"values":    []string{string(jsonBytes)},
		"where":     []string{fmt.Sprintf(`{"user_id":"%d"}`, userID)},
	}, &addResult)

	if err != nil {
		ln.Err(fmt.Sprintf("unable to call apid update method on user_signup_bi"), ln.Map{"err": err.Error()})
		return adaptor.NewError(ErrorDataStorage)
	}

	if addResult != ApidSuccess {
		ln.Err(fmt.Sprintf("apid set for user_signup_bi returned '%s', want 'success'", addResult), nil)
		return adaptor.NewError(ErrorDataStorage)
	}

	return nil
}

// Get package information for signup
func (a *Adaptor) GetSignupPackageInfo(PackageUUID string) client.PackageRecord {
	// define default signup info
	packageInfo := client.PackageRecord{
		ID:             FreePackageID,
		IsFree:         1,
		PackageGroupID: FreePackageGroupID,
		Credits:        FreeAccountCreditsLimits,
	}

	getResult := make([]client.PackageRecord, 0)
	err := a.apidClient.DoFunction("get", url.Values{
		"tableName": []string{"package"},
		"where":     []string{fmt.Sprintf(`{"uuid":"%s"}`, PackageUUID)},
	}, &getResult)
	if err != nil {
		ln.Err("unable to fetch package, we will use default free package", ln.Map{"err": err.Error(), "package_uuid": PackageUUID})
	} else {
		if len(getResult) > 0 && getResult[0].IsFree == 1 {
			packageInfo = getResult[0]
		}
	}

	return packageInfo
}

// Reconstitutes padding removed from a base64 string
func padBase64(s string) string {
	l := len(s)
	m := l % 4
	padding := (4 - m) % 4 // use the modulo so that if m == 0, then padding is 0, not 4

	return s + strings.Repeat("=", padding)
}
