package boss

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/sendgrid/chaos/adaptor"
	redisadaptor "github.com/sendgrid/chaos/adaptor/redis"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/ln"
	"github.com/wunderlist/ttlcache"
)

const (
	UnreachableErrorMessage            = "error reaching boss"
	URLCacheKeyAccountUrl              = "account_url"
	URLCacheKeyAccountCollectURL       = "account_collect_url"
	URLCacheKeyAccountPaymentMethodURL = "account_payment_method_url"
	URLCacheKeyAccountSubscriptionURL  = "account_subscription_url"
	URLCacheKeyPackagePreviewURL       = "package_preview_url"
)

type BossSubscriptionParams struct {
	AccountID       int            `json:"account_id"`
	PackageID       int            `json:"package_id,string"`
	PaymentMethodID string         `json:"payment_method_id"`
	AddOns          []string       `json:"add_ons"`
	Coupon          *client.Coupon `json:"coupon,omitempty"`
}

type BossUpdatePackageParams struct {
	PackageID     int      `json:"package_id"`
	AddOn         []string `json:"add_on"`
	EffectiveDate string   `json:"effective_date"`
}

type BossCancelPackageParams struct {
	EffectiveDate string `json:"effective_date"`
}

var UnreachableError = errors.New(UnreachableErrorMessage)

type Boss interface {
	Subscribe(*BossSubscriptionParams) error
	ChangePackage(userID int, packageID int, immediateChange bool, addOn ...string) *adaptor.AdaptorError
	CancelPackage(userID int) *adaptor.AdaptorError
	CollectBalance(userID int) *adaptor.AdaptorError
	Check() error
	Name() string
}

type Adaptor struct {
	bossURL   string
	authToken string
	urls      urlCache
	keyStore  redisadaptor.KeyStore
}

func New(bossURL string, authToken string, keyStore redisadaptor.KeyStore, urlCacheTTL time.Duration) Boss {
	return &Adaptor{
		bossURL:   bossURL,
		authToken: authToken,
		urls:      NewURLCache(urlCacheTTL, fmt.Sprintf("%s/billing_provider_api/v2/urls", bossURL), authToken),
		keyStore:  keyStore,
	}
}

type BillingProviderURLResponse struct {
	AccountURL              string `json:"account_url"`
	AccountCollectURL       string `json:"account_collect_url"`
	AccountPaymentMethodURL string `json:"account_payment_method_url"`
	AccountSubscriptionURL  string `json:"account_subscription_url"`
	PackagePreviewURL       string `json:"package_preview_url"`
}

type urlCache struct {
	cache              *ttlcache.Cache
	billingProviderURL string
	authToken          string
}

func NewURLCache(ttl time.Duration, billingProviderURL string, authToken string) urlCache {
	u := urlCache{
		cache:              ttlcache.NewCache(ttl),
		billingProviderURL: billingProviderURL,
		authToken:          authToken,
	}

	u.refresh()
	return u
}

func (u *urlCache) accountSubscriptionURL(userID int) (string, error) {
	url, ok := u.cache.Get(URLCacheKeyAccountSubscriptionURL)
	if !ok {
		err := u.refresh()
		if err != nil {
			ln.Err("unable to get boss account subscription url (refresh)", ln.Map{"err": err.Error(), "user_id": userID})
			return "", errors.New("could not refresh cache")
		}
		url, ok = u.cache.Get(URLCacheKeyAccountSubscriptionURL)

		if !ok {
			ln.Err("unable to get boss account subscription url (cache miss)", ln.Map{"err": err.Error(), "user_id": userID})
			return "", errors.New("missing from cache even after refresh")
		}

	}

	return strings.Replace(url, ":account_id", strconv.Itoa(userID), 1), nil
}

func (u *urlCache) accountCollectURL(userID int) (string, error) {
	url, ok := u.cache.Get(URLCacheKeyAccountCollectURL)
	if !ok {
		err := u.refresh()
		if err != nil {
			ln.Err("unable to get boss account collect url (refresh)", ln.Map{"err": err.Error(), "user_id": userID})
			return "", errors.New("could not refresh cache")
		}
		url, ok = u.cache.Get(URLCacheKeyAccountCollectURL)

		if !ok {
			ln.Err("unable to get boss account collect url (cache miss)", ln.Map{"err": err.Error(), "user_id": userID})
			return "", errors.New("missing from cache even after refresh")
		}

	}

	return strings.Replace(url, ":account_id", strconv.Itoa(userID), 1), nil
}

func (u *urlCache) refresh() error {
	req, _ := http.NewRequest("GET", u.billingProviderURL, nil)

	req.Header.Set("Authorization", fmt.Sprintf("token=%s", u.authToken))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		ln.Err("could not refresh billing provider urls cache", ln.Map{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ln.Err("error when getting billing provider urls", ln.Map{"code": resp.StatusCode})
		return errors.New("could not get billing provider urls")
	}

	var urls BillingProviderURLResponse
	err = json.NewDecoder(resp.Body).Decode(&urls)

	u.cache.Set(URLCacheKeyAccountUrl, urls.AccountURL)
	u.cache.Set(URLCacheKeyPackagePreviewURL, urls.PackagePreviewURL)
	u.cache.Set(URLCacheKeyAccountSubscriptionURL, urls.AccountSubscriptionURL)
	u.cache.Set(URLCacheKeyAccountPaymentMethodURL, urls.AccountPaymentMethodURL)
	u.cache.Set(URLCacheKeyAccountCollectURL, urls.AccountCollectURL)

	return nil
}

func (a *Adaptor) Name() string {
	return "boss"
}

func (b *Adaptor) Check() error {
	url := fmt.Sprintf("%s/healthcheck", b.bossURL)
	resp, err := http.Get(url)

	if err != nil {
		ln.Err(
			UnreachableErrorMessage,
			ln.Map{"error": UnreachableError,
				"bossURL": b.bossURL,
			},
		)
		return UnreachableError
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ln.Err(
			UnreachableErrorMessage,
			ln.Map{
				"bossURL": b.bossURL,
			},
		)
		return UnreachableError
	}

	return nil
}

func (b *Adaptor) GetReputationProviderUrl() (string, error) {

	url := fmt.Sprintf("%s/reputation_provider_api/v2/urls", b.bossURL)
	authToken := fmt.Sprintf("token=%s", b.authToken)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	curl := fmt.Sprintf("curl -v -X GET %s --header 'Authorization: <auth token>' --header 'Content-Type: application/json'", url)

	if err != nil {
		return "", fmt.Errorf("http client error (%s) - %s", curl, err)
	}
	defer resp.Body.Close()

	urls, err := simplejson.NewFromReader(resp.Body)

	if err != nil {
		return "", fmt.Errorf("json parse error (%s) - %s", curl, err)
	}

	reportUrl := urls.Get("create_report").MustString("create_report")
	return reportUrl, nil
}

type accountUrlResp struct {
	AccountUrl string `json:"account_url"`
}

func (b *Adaptor) Subscribe(subscription *BossSubscriptionParams) error {
	ln.Debug("Boss Subscribe()", ln.Map{"subscription": subscription})

	authToken := fmt.Sprintf("token=%s", b.authToken)

	client := &http.Client{}

	accountSubscriptionURL, err := b.urls.accountSubscriptionURL(subscription.AccountID)

	if err != nil {
		ln.Err("error when subscribing user, could not get url", ln.Map{"error": err.Error(), "subscription_params": subscription})
		return err
	}

	data, err := json.Marshal(subscription)

	if err != nil {
		ln.Err("something went wrong marshalling subscription", ln.Map{"error": err.Error(), "user_id": subscription.AccountID})
		return err
	}

	req, err := http.NewRequest("POST", accountSubscriptionURL, bytes.NewReader(data))
	if err != nil {
		ln.Err("something went wrong creating http request for boss subscription", ln.Map{"error": err.Error()})
		return err
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	curl := fmt.Sprintf("curl -v -X POST %s -d '%s' --header 'Authorization: <auth token>' --header 'Content-Type: application/json'", accountSubscriptionURL, string(data))
	resp, err := client.Do(req)

	if err != nil {
		ln.Err(fmt.Sprintf("something went wrong executing http request to boss subscription endpoint, %s", curl), ln.Map{"error": err.Error()})
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ln.Err(fmt.Sprintf("Error to read response body, %s", curl), ln.Map{"error": err.Error(), "boss_status_code": resp.StatusCode})
			return err
		}

		ln.Err(fmt.Sprintf("Error posting to subscriptions, %s", curl), ln.Map{"status_code": resp.StatusCode, "response_body": string(b), "user_id": subscription.AccountID, "url": accountSubscriptionURL})

		return errors.New("error when calling internal service")
	}

	return nil
}

func (b *Adaptor) ChangePackage(userID int, packageID int, immediateChange bool, addOn ...string) *adaptor.AdaptorError {
	var effectiveDate time.Time

	if immediateChange {
		effectiveDate = time.Now()
	} else {
		effectiveDate = FindDowngradeDate(time.Now())
	}

	putBody := BossUpdatePackageParams{
		PackageID:     packageID,
		AddOn:         addOn,
		EffectiveDate: effectiveDate.Format("2006-01-02"),
	}

	updatePackageURL := fmt.Sprintf("%s/billing_provider_api/v1/users/%d/user_package", b.bossURL, userID)

	authToken := fmt.Sprintf("token=%s", b.authToken)

	client := &http.Client{}

	data, err := json.Marshal(putBody)

	if err != nil {
		ln.Err("something went wrong marshalling change package put body", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError("could not marshal put body")
	}

	req, err := http.NewRequest("PUT", updatePackageURL, bytes.NewReader(data))
	if err != nil {
		ln.Err("something went wrong creating http request for boss update package", ln.Map{"error": err.Error()})
		return adaptor.NewError("could not create put request")
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	curl := fmt.Sprintf("curl -v -X PUT %s -d '%s' --header 'Authorization: <auth token>' --header 'Content-Type: application/json'", updatePackageURL, string(data))
	resp, err := client.Do(req)

	if err != nil {
		ln.Err(fmt.Sprintf("something went wrong executing http request to boss update user package endpoint, %s", curl), ln.Map{"error": err.Error()})
		return adaptor.NewError("error when calling PUT")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ln.Err(fmt.Sprintf("Error reading response body, %s", curl), ln.Map{"error": err.Error(), "boss_status_code": resp.StatusCode})
			return adaptor.NewError("could not read response body")
		}

		ln.Err(fmt.Sprintf("Error posting to subscriptions, %s", curl), ln.Map{"status_code": resp.StatusCode, "response_body": string(b), "user_id": userID, "url": updatePackageURL})

		return adaptor.NewError("error when calling internal service")
	}

	return nil
}

func (b *Adaptor) CancelPackage(userID int) *adaptor.AdaptorError {
	effectiveDate := FindDowngradeDate(time.Now())

	putBody := BossCancelPackageParams{
		EffectiveDate: effectiveDate.Format("2006-01-02"),
	}
	updatePackageURL := fmt.Sprintf("%s/billing_provider_api/v1/users/%d/user_package/cancel", b.bossURL, userID)

	authToken := fmt.Sprintf("token=%s", b.authToken)

	client := &http.Client{}

	data, err := json.Marshal(putBody)

	if err != nil {
		ln.Err("something went wrong marshalling change package put body", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError("could not marshal put body")
	}

	req, err := http.NewRequest("PUT", updatePackageURL, bytes.NewReader(data))
	if err != nil {
		ln.Err("something went wrong creating http request for boss update package", ln.Map{"error": err.Error()})
		return adaptor.NewError("could not create put request")
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	curl := fmt.Sprintf("curl -v -X PUT %s -d '%s' --header 'Authorization: <auth token>' --header 'Content-Type: application/json'", updatePackageURL, string(data))
	resp, err := client.Do(req)

	if err != nil {
		ln.Err(fmt.Sprintf("something went wrong executing http request to boss update user package endpoint, %s", curl), ln.Map{"error": err.Error()})
		return adaptor.NewError("error when calling PUT")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ln.Err(fmt.Sprintf("Error reading response body, %s", curl), ln.Map{"error": err.Error(), "boss_status_code": resp.StatusCode})
			return adaptor.NewError("could not read response body")
		}

		ln.Err(fmt.Sprintf("Error posting to subscriptions, %s", curl), ln.Map{"status_code": resp.StatusCode, "response_body": string(b), "user_id": userID, "url": updatePackageURL})

		return adaptor.NewError("error when calling internal service")
	}

	return nil
}

func (b *Adaptor) CollectBalance(userID int) *adaptor.AdaptorError {
	ln.Debug("Boss CollectBalance()", ln.Map{"account ID": userID})

	authToken := fmt.Sprintf("token=%s", b.authToken)

	client := &http.Client{}

	accountCollectURL, err := b.urls.accountCollectURL(userID)

	if err != nil {
		ln.Err("error when collecting balance, could not get url", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError("could not get the collect url")
	}

	req, err := http.NewRequest("PUT", accountCollectURL, nil)
	if err != nil {
		ln.Err("something went wrong creating http request for boss collect", ln.Map{"error": err.Error()})
		return adaptor.NewError("could not create http request")
	}
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	curl := fmt.Sprintf("curl -v -X PUT %s --header 'Authorization: <auth token>' --header 'Content-Type: application/json'", accountCollectURL)
	resp, err := client.Do(req)

	if err != nil {
		ln.Err(fmt.Sprintf("something went wrong executing http request to boss collect endpoint, %s", curl), ln.Map{"error": err.Error()})
		return adaptor.NewError("could not execute request to boss collect endpoint")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ln.Err(fmt.Sprintf("Error to read response body, %s", curl), ln.Map{"error": err.Error(), "boss_status_code": resp.StatusCode})
			return adaptor.NewError("could not read response body")
		}

		ln.Err(fmt.Sprintf("Error putting to collect, %s", curl), ln.Map{"status_code": resp.StatusCode, "response_body": string(respBody), "user_id": userID, "url": accountCollectURL})

		return adaptor.NewError("error when calling internal service")
	}

	return nil
}

func FindDowngradeDate(today time.Time) time.Time {
	thisMonth := today.Month()

	var nextMonth time.Month
	year := today.Year()

	if thisMonth == time.December {
		nextMonth = time.January
		year += 1
	} else {
		nextMonth = thisMonth + 1
	}

	return time.Date(year, nextMonth, 1, 0, 0, 0, 0, time.UTC)

}
