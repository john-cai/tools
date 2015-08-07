package apidadaptor

import (
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/ln"
)

const (
	FirstIPPolicy = "first_ip"
)

type IPService interface {
	ValidateIPs(int, []string) (bool, *adaptor.AdaptorError)
	AddUserSendIP(int, string) (int, *adaptor.AdaptorError)
	UnassignExternalIps([]int) (int, *adaptor.AdaptorError)
	DeleteUserIp(int, string) (int, *adaptor.AdaptorError)
	DeleteAllUserIps([]int) (int, *adaptor.AdaptorError)
	DeleteAllUserIpGroups([]int) (int, *adaptor.AdaptorError)
	CountExternalIP(int) (int, *adaptor.AdaptorError)
	DeleteUserIPGroup(int, int) *adaptor.AdaptorError
	AssignFirstIP(int) *adaptor.AdaptorError
	GetFirstIP() (string, *adaptor.AdaptorError)
	AssignExternalIP(int, string) *adaptor.AdaptorError
	SetUserIPGroup(int, int, int) *adaptor.AdaptorError
	GetUserSendIps(int) ([]string, *adaptor.AdaptorError)
}

type ExternalIP struct {
	IP            string `json:"ip"`
	ResellerID    int    `json:"reseller_id"`
	ServerID      int    `json:"server_name_id"`
	InSenderScore int    `json:"in_sender_score"`
}

// Get the first available IP and assign it to the user.  Set that IP as the user send IP
func (a *Adaptor) AssignFirstIP(userID int) *adaptor.AdaptorError {
	locations, err := a.getFirstIPLocation()

	if len(locations) == 0 {
		ln.Info("no locations found for first_ip policy", ln.Map{"locations": locations})
		return adaptor.NewError("no locations found for first_ip policy")
	}

	if err != nil {
		ln.Err("error when getting server locations", ln.Map{"error": err.Error()})
		return adaptor.NewError("error when getting server locations")
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))

	params := url.Values{
		"userid":          []string{strconv.Itoa(userID)},
		"limit":           []string{strconv.Itoa(1)},
		"server_location": []string{strconv.Itoa(locations[r.Intn(len(locations))])},
	}

	// Grab the first available IP and immediately assign it to the user
	var IP []string
	apidErr := a.apidClient.DoFunction("assignBestAvailableOp", params, &IP)
	if apidErr != nil {
		ln.Err("error assigning best ips", ln.Map{"error": apidErr.Error(), "params": params})
		return adaptor.NewError("error assigning best available ips")
	}

	if len(IP) == 0 {
		ln.Info("no available ips", ln.Map{"locations": locations})
		return adaptor.NewError("no available ips")
	}

	// assign ip to user send ip
	_, err = a.AddUserSendIP(userID, IP[0])
	if err != nil {
		ln.Err("could not add ip to user send ips table", ln.Map{"err": err.Error(), "user_id": userID})
		return adaptor.NewError("could not add ip to user send ips table")
	}

	return nil
}

// GetFirstIP finds an ip from the location based on the first_ip policy
func (a *Adaptor) GetFirstIP() (string, *adaptor.AdaptorError) {

	locations, err := a.getFirstIPLocation()

	if len(locations) == 0 {
		ln.Info("no locations found for first_ip policy", ln.Map{"locations": locations})
		return "", adaptor.NewError("no locations found for first_ip policy")
	}

	if err != nil {
		ln.Err("error when getting server locations", ln.Map{"error": err.Error()})
		return "", adaptor.NewError("error when getting server locations")
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))

	params := url.Values{
		"limit":           []string{strconv.Itoa(1)},
		"server_location": []string{strconv.Itoa(locations[r.Intn(len(locations))])},
	}

	var IP []string
	getIPErr := a.apidClient.DoFunction("getBestAvailableIp", params, &IP)
	if getIPErr != nil {
		ln.Err("error getting best ips", ln.Map{"error": getIPErr.Error(), "params": params})
		return "", adaptor.NewError("error getting best available ips")
	}

	if len(IP) == 0 {
		ln.Info("no avaiable ips", ln.Map{"locations": locations})
		return "", adaptor.NewError("no available ips")
	}

	return IP[0], nil

}

func (a *Adaptor) AssignExternalIP(userID int, ip string) *adaptor.AdaptorError {

	var eIP []ExternalIP

	err := a.apidClient.DoFunction("getExternalIp", url.Values{"ip": []string{ip}, "exclude_whitelabels": []string{strconv.Itoa(0)}}, &eIP)

	if err != nil {
		ln.Err("error when getting external ips of user", ln.Map{"error": err.Error(), "user_id": userID, "ip": ip})
		return adaptor.NewError("error when getting external ips of user")

	}
	params := url.Values{
		"ip":              []string{ip},
		"reseller_id":     []string{strconv.Itoa(userID)},
		"server_name_id":  []string{strconv.Itoa(eIP[0].ServerID)},
		"in_sender_score": []string{strconv.Itoa(eIP[0].InSenderScore)},
	}

	var result int
	err = a.apidClient.DoFunction("editExternalIp", params, &result)
	if err != nil {
		ln.Err("error when editing external ip of user", ln.Map{"error": err.Error(), "user_id": userID, "ip": ip})
		return adaptor.NewError("error when editing external ip of user")
	}

	return nil

}

func (a *Adaptor) CountExternalIP(userID int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
	}

	var countIps int
	err := a.apidClient.DoFunction("countExternalIp", params, &countIps)
	if err != nil {
		ln.Err("error retrieving external IPs count", ln.Map{"error": err.Error(), "user_id": userID})
		return 0, adaptor.NewError("error retrieving external IPs count")
	}

	return countIps, nil
}

func (a *Adaptor) SetUserIPGroup(userID int, groupID int, oldGroupID int) *adaptor.AdaptorError {
	params := url.Values{
		"userid":     []string{strconv.Itoa(userID)},
		"groupid":    []string{strconv.Itoa(groupID)},
		"oldgroupid": []string{strconv.Itoa(oldGroupID)},
	}
	var result int

	err := a.apidClient.DoFunction("setUserIpGroup", params, &result)
	if err != nil {
		ln.Err("error when setting user ip group", ln.Map{"error": err.Error(), "user_id": userID, "group_id": groupID, "old_group_id": oldGroupID})
		return adaptor.NewError("error when setting user ip group")
	}

	return nil

}

func (a *Adaptor) DeleteUserIPGroup(userID int, groupID int) *adaptor.AdaptorError {
	params := url.Values{
		"userid":  []string{strconv.Itoa(userID)},
		"groupid": []string{strconv.Itoa(groupID)},
	}
	var result int

	err := a.apidClient.DoFunction("removeUserIpGroup", params, &result)
	if err != nil {
		ln.Err("error when deleting user ip group", ln.Map{"error": err.Error(), "user_id": userID, "group_id": groupID})
		return adaptor.NewError("error when deleting user ip group")
	}

	return nil

}

// getFirstIPLocation finds a list of location_ids based on the "first_ip" policy
func (a *Adaptor) getFirstIPLocation() ([]int, *adaptor.AdaptorError) {
	var locations []int

	params := url.Values{
		"policy": []string{FirstIPPolicy},
	}

	err := a.apidClient.DoFunction("getAssignmentPolicy", params, &locations)
	if err != nil {
		ln.Err("error when getting assignment policy", ln.Map{"error": err.Error()})
		return []int{}, adaptor.NewError("error when getting first ip location")
	}

	return locations, nil
}

func (a *Adaptor) ValidateIPs(userID int, ips []string) (bool, *adaptor.AdaptorError) {
	var formattedErr *adaptor.AdaptorError
	if len(ips) < 1 {
		formattedErr = adaptor.NewErrorWithStatus("No ips provided", http.StatusBadRequest)
		return false, formattedErr
	}

	params := url.Values{
		"userid":  []string{strconv.Itoa(userID)},
		"ip_list": ips,
	}

	var validIPs int
	err := a.apidClient.DoFunction("validateExternalIps", params, &validIPs)

	if err != nil {
		formattedErr = adaptor.NewError(err.Error())
		if strings.Contains(err.Error(), "ips were invalid") {
			formattedErr = adaptor.NewErrorWithStatus("One or more ips were invalid", http.StatusBadRequest)
		}
		return false, formattedErr
	}

	if validIPs < 1 {
		formattedErr = adaptor.NewErrorWithStatus("User does not have any IPs", http.StatusNotFound)
		return false, formattedErr
	}

	return true, nil
}

func (a *Adaptor) AddUserSendIP(userID int, IP string) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
		"ip":     []string{IP},
	}

	var count int
	err := a.apidClient.DoFunction("addUserSendIp", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) UnassignExternalIps(userIDs []int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userids": toStringSlice(userIDs),
	}

	var count int
	err := a.apidClient.DoFunction("unassignExternalIps", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) DeleteUserIp(userID int, ip string) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
		"ip":     []string{ip},
	}

	var count int
	err := a.apidClient.DoFunction("deleteUserSendIp", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) DeleteAllUserIps(userIDs []int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userids": toStringSlice(userIDs),
	}

	var count int
	err := a.apidClient.DoFunction("deleteAllUserIps", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) DeleteAllUserIpGroups(userIDs []int) (int, *adaptor.AdaptorError) {
	params := url.Values{
		"userids": toStringSlice(userIDs),
	}

	var count int
	err := a.apidClient.DoFunction("deleteAllUserIpGroups", params, &count)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return 0, formattedErr
	}

	return count, nil
}

func (a *Adaptor) GetUserSendIps(userID int) ([]string, *adaptor.AdaptorError) {
	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
	}

	var ips []string
	err := a.apidClient.DoFunction("getUserSendIp", params, &ips)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return []string{}, formattedErr
	}

	return ips, nil
}
