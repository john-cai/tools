package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"code.google.com/p/go-uuid/uuid"

	apidadaptor "github.com/sendgrid/chaos/adaptor/apid"
	"github.com/sendgrid/go-apid"
)

var TotalUsers, SubusersPerUser int
var ChaosUrl, ApidUrl string
var ApidAdaptor *apidadaptor.Adaptor
var ApidClient *apid.HTTPClient
var ChaosPort = 50110

const (
	IPGroupFree = 1
)

func main() {
	flag.IntVar(&TotalUsers, "users", 1, "number of users")
	flag.IntVar(&SubusersPerUser, "subusers", 1, "number of subusers to create per user")
	flag.StringVar(&ChaosUrl, "chaos", "localhost", "chaos url")
	flag.StringVar(&ApidUrl, "apid", "localhost", "apid url")
	flag.Parse()

	ApidClient = apid.NewHTTPClient(fmt.Sprintf("ApidUrl:%d", 8082))
	ApidAdaptor = apidadaptor.New(ApidClient)

	var wg sync.WaitGroup

	for i := 0; i < TotalUsers; i++ {
		wg.Add(1)
		go func() {
			userID := createUserAndAssignIP()
			createSubusers(userID)
			wg.Done()
		}()
	}

	wg.Wait()
}

type SignupResponse struct {
	Username         string            `json:"username"`
	UserID           int               `json:"user_id"`
	Email            string            `json:"email"`
	SGToken          string            `json:"signup_session_token"`
	Token            string            `json:"authorization_token"`
	CreditAllocation *CreditAllocation `json:"credit_allocation,omitempty"`
}

type CreditAllocation struct {
	Type CreditAllocationType `json:"type"`
}

type CreditAllocationType string

func createUserAndAssignIP() int {
	resp, err := createUser()
	fmt.Println("user created!")

	if err != nil {
		fmt.Printf("oh no there's an error! %s\n", err.Error())
	}
	//set user package
	err = ApidAdaptor.DeleteUserIPGroup(resp.UserID, IPGroupFree)
	if err != nil {
		return 0
	}

	ApidAdaptor.SetUserPackage(resp.UserID, 11)

	// get the first available IP and immediately assign it to the user
	adaptorErr := ApidAdaptor.AssignFirstIP(resp.UserID)
	if adaptorErr != nil {
		return 0
	}

	return resp.UserID

}

var SteadfastLocationId = 5

func generateRandomIP() string {
	ip := fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256))
	return ip
}

func generateNewIP() error {

	defaultLocation := SteadfastLocationId
	var serverNameId int
	testIP := generateRandomIP()

	// the process of selecting an ip for the user is as follows:
	// 1. get a list of locations from the ip_assignment_policy table
	// 2. get an ip based from those locations
	// we need to add a server for a given location, as well as an external ip to live on that server so
	// we can "find" that ip to assign
	var newServerLocationID int
	ApidClient.DoFunction("executeSql", url.Values{
		"query":    []string{fmt.Sprintf(`insert into server_location (id,name) values(%d,"test")`, defaultLocation)},
		"rw":       []string{"1"},
		"resource": []string{"mail"},
		"insert":   []string{"1"},
	}, &newServerLocationID)

	ApidClient.DoFunction("addServerName", url.Values{
		"ip":       []string{generateRandomIP()},
		"server":   []string{"testservername"},
		"type":     []string{"proxy"},
		"location": []string{strconv.Itoa(defaultLocation)},
	}, &serverNameId)

	var setPolicySuccess int
	ApidClient.DoFunction("addAssignmentPolicy", url.Values{
		"policy":   []string{"first_ip"},
		"location": []string{strconv.Itoa(defaultLocation)},
	}, setPolicySuccess)

	var addExternalIpSuccess int
	addIPErr := ApidClient.DoFunction("addExternalIp", url.Values{
		"ip":             []string{testIP},
		"server_name_id": []string{strconv.Itoa(serverNameId)},
	}, &addExternalIpSuccess)
	if addIPErr != nil && addIPErr.Error() != `"key exists"` {
		return addIPErr
	}

	return nil
}

func createSubusers(resellerID int) error {
	/*	createSubuserURL := fmt.Sprintf("http://%s:%d/v1/users/%d/subusers", ChaosUrl, ChaosPort, resellerID)

		return nil
	*/
	return nil
}

func createUser() (SignupResponse, error) {
	username := fmt.Sprintf("testuser_%s", uuid.New())
	email := fmt.Sprintf("testuser_%s@sendgrid.com", uuid.New())
	password := "very secure password"
	return createSpecificUser(username, email, password)
}

// createUser is a helper method to create a user assuming username, email, and password are valid
func createSpecificUser(username string, email string, password string) (SignupResponse, error) {
	createUserURL := fmt.Sprintf("http://%s:%d/v1/signup", ChaosUrl, ChaosPort)
	var jsonData = []byte(fmt.Sprintf(`{"username":"%s", "email":"%s", "password":"%s"}`, username, email, password))
	var resp SignupResponse

	req, err := http.NewRequest("POST", createUserURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return resp, err
	}
	req.Header.Set("Content-Type", `application/json`)
	req.Header.Set("X-Mako", `{"ip":"192.168.1.700"}`)

	httpClient := &http.Client{}
	createResponse, err := httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer createResponse.Body.Close()

	if createResponse.StatusCode != http.StatusCreated {
		b, _ := ioutil.ReadAll(createResponse.Body)
		errStr := "got status %d creating user in helper function. response: %s "
		errStr += "curl -v %s -d '%s'"
		return resp, fmt.Errorf(errStr, createResponse.StatusCode, string(b), createUserURL, string(jsonData))
	}

	err = json.NewDecoder(createResponse.Body).Decode(&resp)
	if err != nil {
		return resp, err
	}

	if resp.UserID == 0 {
		errStr := "unhandled error creating new user in helper function. "
		errStr += "curl -v %s -d '%s'; got status: %d, No ID. %#v"
		return resp, fmt.Errorf(errStr, createUserURL, string(jsonData), createResponse.StatusCode, resp)
	}

	err = setUserToActive(resp.UserID)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func setUserToActive(userID int) error {
	ApidClient := apid.NewHTTPClient(ApidUrl)

	params := url.Values{
		"userid": []string{strconv.Itoa(userID)},
		"active": []string{strconv.Itoa(1)},
	}

	var updated int
	err := ApidClient.DoFunction("setUserActive", params, &updated)
	if err != nil {
		return errors.New("unable to activate parent")
	}

	return nil
}
