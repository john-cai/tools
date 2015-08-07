package authzd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sendgrid/ln"
)

type Adaptor struct {
	Host            string
	Port            int
	HealthcheckPort int
}

type Authorizer interface {
	SetUserTemplate(userID int, template string, removeScopes ...string) error
	CreateUserScopeSet(userID int, template string) (string, error)
	VerifyScopes(scopeSetID string, scopes ...string) (map[string]bool, error)
	RequireScopes(scopeSetID string, scopes ...string) (bool, error)
	HasTemplateScope(templateName string, findScope string) (bool, error)
}

type ScopeTemplate struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type scopeSetID struct {
	ID string `json:"scope_set_id"`
}

const (
	TemplateConfirmEmailPage = "confirm_email_page"
	TemplateProvisionPage    = "provision_page"
)

func New(host string, port int, healthcheckPort int) *Adaptor {
	a := &Adaptor{
		Host:            host,
		Port:            port,
		HealthcheckPort: healthcheckPort,
	}

	return a
}

// Name adheres to the adminiable interface
func (adaptor *Adaptor) Name() string {
	return "authzd"
}

// Check adheres to the adminiable interface and
// provides the way to check that this adaptor is working correctly
func (adaptor *Adaptor) Check() error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/healthcheck", adaptor.Host, adaptor.HealthcheckPort))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code - got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	return nil
}

// SetUserTemplate sets the user's scope set using the scope set template
// any removeScopes will be removed
func (adaptor *Adaptor) SetUserTemplate(userID int, template string, removeScopes ...string) error {
	putURL := fmt.Sprintf("http://%s:%d/v1/permissions/users/%d/scopeset", adaptor.Host, adaptor.Port, userID)

	data := fmt.Sprintf(`{"new_template":"%s"}`, template)
	if len(removeScopes) > 1 {
		for i, remove := range removeScopes {
			removeScopes[i] = fmt.Sprintf(`"%s"`, remove)
		}
		removePermissions := strings.Join(removeScopes, ",")
		data = fmt.Sprintf(`{"new_template":"%s","remove_permissions":[%s]}`, template, removePermissions)
	}
	putBody := bytes.NewBufferString(data)

	req, err := http.NewRequest("PUT", putURL, putBody)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling %s - body %s - err %s", putURL, putBody, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// disregard error as body only helps to give a more detailed err msg
		errBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error response from %s - body %s - status %d", putURL, errBody, resp.StatusCode)
	}

	return nil
}

// CreateUserScopeSet creates a scope set for a given user using a scope set template, returns scope_set_id
func (adaptor *Adaptor) CreateUserScopeSet(userID int, template string) (string, error) {
	postURL := fmt.Sprintf("http://%s:%d/v1/permissions/users/%d/scopeset", adaptor.Host, adaptor.Port, userID)

	data := bytes.NewBufferString(fmt.Sprintf(`{"new_template":"%s"}`, template))
	dataBackup := bytes.NewBufferString(fmt.Sprintf(`{"new_template":"%s"}`, template))
	resp, err := http.Post(postURL, "application/json", data)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"method": "SetUserTemplate", "user_id": userID, "template": template})
		return "", fmt.Errorf("error posting url %s - data %s - err %s", postURL, data, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		ln.Err("authzd adaptor error", ln.Map{"method": "SetUserTemplate", "reason": fmt.Sprintf("error posting url %s - data %s - got status %d, want %d; body %s", postURL, dataBackup, resp.StatusCode, http.StatusCreated, body)})
		return "", fmt.Errorf("error posting url %s - data %s - got status %d, want %d; body %s", postURL, dataBackup, resp.StatusCode, http.StatusCreated, body)
	}
	var result scopeSetID

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return "", fmt.Errorf("error getting scope set id on create user scope set call")
	}

	return result.ID, nil

}

//GetUserScopeSetID returns the user's scope set id (uuid)
func (adaptor *Adaptor) GetUserScopeSetID(userID int) (string, error) {
	getURL := fmt.Sprintf("http://%s:%d/v1/permissions/users/%d/scopeset", adaptor.Host, adaptor.Port, userID)

	resp, err := http.Get(getURL)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "user_id": userID})
		return "", fmt.Errorf("error getting url %s - %s", getURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("error getting url %s - got status %d, want %d; %s", getURL, resp.StatusCode, http.StatusOK, body)
	}

	data := scopeSetID{}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "method": "GetUserScopeSetID", "user_id": userID})
		return "", fmt.Errorf("error decoding body from url %s, %s", getURL, err)
	}

	return data.ID, nil
}

// GetCredentialScopeSetID returns the credentials's scope set id (uuid)
func (adaptor *Adaptor) GetCredentialScopeSetID(credentialID int) (string, error) {
	getURL := fmt.Sprintf("http://%s:%d/v1/permissions/credentials/%d/scopeset", adaptor.Host, adaptor.Port, credentialID)

	resp, err := http.Get(getURL)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "credential_id": credentialID})
		return "", fmt.Errorf("error getting url %s - %s", getURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("error getting url %s - got status %d, want %d; %s", getURL, resp.StatusCode, http.StatusOK, body)
	}

	data := scopeSetID{}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "method": "GetCredentialScopeSetID", "credential_id": credentialID})
		return "", fmt.Errorf("error decoding body from url %s, %s", getURL, err)
	}

	return data.ID, nil
}

// AddTemplate inserts a template record
func (adaptor *Adaptor) AddTemplate(scopeTemplate ScopeTemplate) error {
	postURL := fmt.Sprintf("http://%s:%d/v1/permissions/scopesettemplates", adaptor.Host, adaptor.Port)
	b, err := json.Marshal(scopeTemplate)

	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "method": "AddTemplate", "scopeTemplate": scopeTemplate})
		return fmt.Errorf("error marshalling scope template - %s", err)
	}
	data := bytes.NewBuffer(b)

	resp, err := http.Post(postURL, "application/json", data)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "method": "AddTemplate", "scopeTemplate": scopeTemplate})
		return fmt.Errorf("error posting data '%s' to url %s", b, postURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := ioutil.ReadAll(resp.Body)
		ln.Err("authzd adaptor error", ln.Map{"error": fmt.Sprintf("error posting data '%s' to url %s - got status %d, want 200 or 409; %s", b, postURL, resp.StatusCode, body), "method": "AddTemplate", "scopeTemplate": scopeTemplate})
		return fmt.Errorf("error posting data '%s' to url %s - got status %d, want 200 or 409; %s", b, postURL, resp.StatusCode, body)

	}

	return nil
}

// HasTemplateScope returns true if the given template contains the desired scope
func (adaptor *Adaptor) HasTemplateScope(templateName string, findScope string) (bool, error) {
	var scopeTemplate ScopeTemplate

	getURL := fmt.Sprintf("http://%s:%d/v1/permissions/scopesettemplates?template_name=%s", adaptor.Host, adaptor.Port, url.QueryEscape(templateName))

	resp, err := http.Get(getURL)
	if err != nil {
		return false, fmt.Errorf("error getting url %s - %s", getURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("error getting url %s - got status %d, want %d; %s", getURL, resp.StatusCode, http.StatusOK, body)
	}

	err = json.NewDecoder(resp.Body).Decode(&scopeTemplate)
	if err != nil {
		return false, fmt.Errorf("error decoding body from url %s, %s", getURL, err)
	}

	for _, scope := range scopeTemplate.Scopes {
		if scope == findScope {
			return true, nil
		}
	}

	return false, nil
}

func (adaptor *Adaptor) DeleteCredentialScopeSet(credentialID int) error {
	postURL := fmt.Sprintf("http://%s:%d/v1/permissions/credentials/%d/scopeset", adaptor.Host, adaptor.Port, credentialID)
	req, _ := http.NewRequest("DELETE", postURL, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ln.Err("authzd adaptor error", ln.Map{"error": err.Error(), "method": "DeleteCredentialScopeSet", "credential_id": credentialID})
		return fmt.Errorf("error deleting credential scope set for credential %d", credentialID)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		ln.Err("authzd adaptor error", ln.Map{"error": fmt.Sprintf("error deleting data '%s' to url %s - got status %d, want 200 ; %s", postURL, resp.StatusCode, body), "method": "DeleteCredentialScopeSet", "credential_id": credentialID})
		return fmt.Errorf("error posting to url %s - got status %d, want 200; %s", postURL, resp.StatusCode, body)
	}

	return nil
}

// VerifyScopes returns true/false mapping for each scope
// This method is "greedy" - it will return more keys than requested
// based on like% matching
func (adaptor *Adaptor) VerifyScopes(scopeSetID string, scopes ...string) (map[string]bool, error) {
	data := make(map[string]bool)

	scopeQueryParams := make([]string, len(scopes))
	for i, _ := range scopes {
		scopeQueryParams[i] = "scopes=" + scopes[i]
	}

	getURL := fmt.Sprintf("http://%s:%d/v1/permissions/scopesets/%s/verify?%s", adaptor.Host, adaptor.Port, scopeSetID, strings.Join(scopeQueryParams, "&"))

	resp, err := http.Get(getURL)
	if err != nil {
		return data, fmt.Errorf("error getting url %s - %s", getURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return data, fmt.Errorf("error getting url %s - got status %d, want %d; %s", getURL, resp.StatusCode, http.StatusOK, body)
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, fmt.Errorf("error decoding body from url %s, %s", getURL, err)
	}

	return data, nil
}

func (adaptor *Adaptor) RequireScopes(scopeSetID string, scopes ...string) (bool, error) {
	results, err := adaptor.VerifyScopes(scopeSetID, scopes...)
	if err != nil {
		return false, err
	}

	for _, value := range results {
		if value == false {
			return false, nil
		}
	}

	return true, nil
}

// Error is passed back to callers
type AdaptorError struct {
	Err                 error
	SuggestedStatusCode int
}

// Error matches the interface for errors
func (e *AdaptorError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// NewError creates a default adaptor error with http.StatusInternalServerError
func NewError(msg string) *AdaptorError {
	return &AdaptorError{
		Err:                 errors.New(msg),
		SuggestedStatusCode: http.StatusInternalServerError,
	}
}

// NewErrorWithStatus lets you specify the status code
func NewErrorWithStatus(msg string, status int) *AdaptorError {
	return &AdaptorError{
		Err:                 errors.New(msg),
		SuggestedStatusCode: status,
	}
}
