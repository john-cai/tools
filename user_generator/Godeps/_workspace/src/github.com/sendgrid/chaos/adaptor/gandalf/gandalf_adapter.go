package gandalf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sendgrid/ln"
)

const (
	ErrorUnreachable    = "error reaching gandalf"
	ErrorAuthentication = "error authenticating"
	ErrorBadCredentials = "error authenticating due to bad credentials"
	ErrorChangePassword = "error updating password"
)

var unreachable = errors.New(ErrorUnreachable)
var getTokenError = errors.New("Error when getting user token")

type Gandalf interface {
	GetAuthorizationToken(userid int) (string, error)
	ExpireTokens(userid int) error
	ValidateToken(token string) (int, error)
	ValidatePassword(username string, passwords string) (*PasswordValidationResponse, error)
	ChangePassword(userID int, newPassword string) error
	Check() error
}

type Adaptor struct {
	GandalfHost            string
	GandalfPort            int
	GandalfHealthcheckPort int
}

type ValidateRequest struct {
	Token string `json:"token"`
}
type TokenResponse struct {
	Token  string `json:"token,omitempty"`
	UserId int    `json:"user_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

func New(host string, port int, healthcheckPort int) *Adaptor {
	return &Adaptor{
		GandalfHost:            host,
		GandalfPort:            port,
		GandalfHealthcheckPort: healthcheckPort,
	}
}

func (g *Adaptor) Check() error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/healthcheck", g.GandalfHost, g.GandalfHealthcheckPort))
	if err != nil {
		ln.Err(
			ErrorUnreachable,
			ln.Map{
				"error":                    err.Error(),
				"gandalf_host":             g.GandalfHost,
				"gandalf_healthcheck_port": g.GandalfHealthcheckPort,
			},
		)
		return unreachable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ln.Err(
			ErrorUnreachable,
			ln.Map{
				"gandalf_host":             g.GandalfHost,
				"gandalf_healthcheck_port": g.GandalfHealthcheckPort,
			},
		)
		return unreachable
	}
	return nil
}

func (g *Adaptor) Name() string {
	return "gandalf"
}

type PasswordValidationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PasswordValidationResponse struct {
	UserID     int    `json:"user_id"`
	RemoteAddr string `json:"remote_addr"`
	ScopeSetID string `json:"scope_set_id"`
	Error      string `json:"error"`
}

func (g *Adaptor) ValidatePassword(username string, password string) (*PasswordValidationResponse, error) {
	authenticationURL := fmt.Sprintf("http://%s:%d/validate/password", g.GandalfHost, g.GandalfPort)

	authenticationReq := PasswordValidationRequest{
		Username: username,
		Password: password,
	}

	data, err := json.Marshal(authenticationReq)

	if err != nil {
		ln.Err("could not marshal authentication request parameters", ln.Map{"error": err.Error(), "username": username})
		return nil, errors.New(ErrorAuthentication)
	}

	req, err := http.NewRequest("POST", authenticationURL, bytes.NewBuffer(data))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		ln.Err("error posting to gandalf authentication endpoint", ln.Map{"error": err.Error(), "username": username})
		return nil, errors.New(ErrorAuthentication)
	}

	defer resp.Body.Close()
	var authenticationResp PasswordValidationResponse

	err = json.NewDecoder(resp.Body).Decode(&authenticationResp)

	if err != nil {
		ln.Err("error decoding response from gandalf", ln.Map{"error": err.Error(), "username": username})
		return nil, errors.New(ErrorAuthentication)
	}

	if authenticationResp.Error != "" {
		if authenticationResp.Error == "Unauthorized" {
			ln.Info("authentication failed because of bad credentials", ln.Map{"username": username})
			return nil, errors.New(ErrorBadCredentials)
		}
		ln.Err("authentication failed", ln.Map{"username": username, "error": authenticationResp.Error})
		return nil, errors.New(ErrorAuthentication)
	}

	return &authenticationResp, nil
}

type ChangePasswordRequest struct {
	UserID   int    `json:"user_id"`
	Password string `json:"password"`
}

func (g *Adaptor) ChangePassword(userID int, password string) error {

	changePasswordURL := fmt.Sprintf("http://%s:%d/password", g.GandalfHost, g.GandalfPort)

	changePWReq := &ChangePasswordRequest{
		UserID:   userID,
		Password: password,
	}

	data, err := json.Marshal(changePWReq)

	if err != nil {
		ln.Err("could not marshal changePassword request parameters", ln.Map{"error": err.Error(), "user_id": userID})
		return errors.New(ErrorChangePassword)
	}

	req, err := http.NewRequest("PUT", changePasswordURL, bytes.NewBuffer(data))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		ln.Err("error posting to gandalf password endpoint", ln.Map{"error": err.Error(), "user_id": userID})
		return errors.New(ErrorChangePassword)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ln.Err("error calling password endpoint in gandalf", ln.Map{"status": resp.StatusCode, "user_id": userID})
		return errors.New(ErrorChangePassword)
	}

	return nil
}

func (g *Adaptor) ValidateToken(token string) (int, error) {

	request := ValidateRequest{Token: token}

	gandalfUrl := fmt.Sprintf("http://%s:%d/validate", g.GandalfHost, g.GandalfPort)
	client := http.Client{}

	marshalledReq, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", gandalfUrl, bytes.NewBuffer(marshalledReq))
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}

	var validateResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&validateResponse)
	if err != nil {
		return 0, err
	}

	if validateResponse.Error != "" {
		return 0, fmt.Errorf("validation failed: %s", validateResponse.Error)
	}

	return validateResponse.UserId, nil

}

type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (g *Adaptor) ExpireTokens(userID int) error {
	request := ExpireTokenRequest{Action: "all"}

	gandalfUrl := fmt.Sprintf("http://%s:%d/expire/%d", g.GandalfHost, g.GandalfPort, userID)
	client := http.Client{}

	marshalledReq, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", gandalfUrl, bytes.NewBuffer(marshalledReq))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to expire tokens")
	}

	return nil
}

type ExpireTokenRequest struct {
	Action string `json:"action"`
}

type gandalfToken struct {
	UserID     int    `json:"user_id"`
	RemoteAddr string `json:"remote_addr"`
	Token      string `json:"token"`
	ScopeSetID string `json:"scope_set_id"`
}

func (g *Adaptor) GetAuthorizationToken(userid int) (string, error) {
	// Get existing token if it exists
	getURL := fmt.Sprintf("http://%s:%d/tokens/%d", g.GandalfHost, g.GandalfPort, userid)

	resp, err := http.Get(getURL)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokens []gandalfToken
	err = json.NewDecoder(resp.Body).Decode(&tokens)

	if err != nil {
		ln.Err("error when decoding tokens", ln.Map{"method": "GetAuthorizationToken", "user_id": userid, "error": err.Error()})
	}

	if len(tokens) > 0 {
		return tokens[0].Token, nil
	}

	generateURL := fmt.Sprintf("http://%s:%d/generate/%d", g.GandalfHost, g.GandalfPort, userid)
	client := http.Client{}

	req, err := http.NewRequest("GET", generateURL, nil)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		ln.Err("error when create request for the generate token url", ln.Map{"method": "GetAuthorizationToken", "user_id": userid, "error": err.Error()})
		return "", getTokenError
	}

	resp, err = client.Do(req)
	if err != nil {
		ln.Err("error when calling generate token", ln.Map{"method": "GetAuthorizationToken", "user_id": userid, "error": err.Error()})
		return "", getTokenError
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		ln.Err("error when decoding gnereated token response", ln.Map{"method": "GetAuthorizationToken", "user_id": userid, "error": err.Error()})
		return "", getTokenError
	}

	return tokenResponse.Token, nil

}
