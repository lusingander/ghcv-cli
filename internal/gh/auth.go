package gh

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	oauthClientId  = "8f2a9bd8ba029f2a0e17"
	deviceCodeUrl  = "https://github.com/login/device/code"
	accessTokenUrl = "https://github.com/login/oauth/access_token"
	scope          = "" // read-only access to public information
	grantType      = "urn:ietf:params:oauth:grant-type:device_code"
)

func postLogin(url string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

func (r *deviceCodeResponse) valid() bool {
	return r.DeviceCode != ""
}

func postDeviceCode() (*deviceCodeResponse, error) {
	values := url.Values{}
	values.Add("client_id", oauthClientId)
	values.Add("scope", scope)

	body, err := postLogin(deviceCodeUrl, values)
	if err != nil {
		return nil, err
	}

	res := &deviceCodeResponse{}
	err = json.Unmarshal(body, res)
	if err == nil && res.valid() {
		return res, nil
	}
	return nil, err
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func (r *accessTokenResponse) valid() bool {
	return r.AccessToken != ""
}

type accessTokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

func (r *accessTokenErrorResponse) valid() bool {
	return r.Error != ""
}

func (r *accessTokenErrorResponse) toError() error {
	return fmt.Errorf("%s %s %s", r.Error, r.ErrorDescription, r.ErrorUri)
}

func postAccessToken(deviceCode string) (*accessTokenResponse, *accessTokenErrorResponse, error) {
	values := url.Values{}
	values.Add("client_id", oauthClientId)
	values.Add("device_code", deviceCode)
	values.Add("grant_type", grantType)

	body, err := postLogin(accessTokenUrl, values)
	if err != nil {
		return nil, nil, err
	}

	res := &accessTokenResponse{}
	err = json.Unmarshal(body, res)
	if err == nil && res.valid() {
		return res, nil, nil
	}

	errRes := &accessTokenErrorResponse{}
	err = json.Unmarshal(body, errRes)
	if err == nil && errRes.valid() {
		return nil, errRes, nil
	}

	return nil, nil, err
}

func pollAccessToken(r *deviceCodeResponse) (*accessTokenResponse, error) {
	deviceCode := r.DeviceCode
	interval := time.Duration(r.Interval+1) * time.Second
	for {
		time.Sleep(interval)
		acResp, acErrResp, err := postAccessToken(deviceCode)
		if err != nil {
			return nil, err
		}
		if acErrResp != nil {
			switch acErrResp.Error {
			case "authorization_pending":
				continue
			case "slow_down":
				interval *= 2
				continue
			default:
				return nil, acErrResp.toError()
			}
		}
		return acResp, nil
	}
}

func openBrowser(url string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("unsupported os")
	}
	cmd := exec.Command("open", url)
	return cmd.Start()
}

func promptInputUserCode(r *deviceCodeResponse) error {
	fmt.Println("Enter this code:", r.UserCode)
	fmt.Println(r.VerificationURI)
	return openBrowser(r.VerificationURI)
}

func authDeviceFlow() (string, error) {
	// https://docs.github.com/ja/developers/apps/building-oauth-apps/authorizing-oauth-apps#device-flow
	dcResp, err := postDeviceCode()
	if err != nil {
		return "", err
	}
	if err := promptInputUserCode(dcResp); err != nil {
		return "", err
	}
	atResp, err := pollAccessToken(dcResp)
	if err != nil {
		return "", err
	}
	return atResp.AccessToken, nil
}

func Authorize() (*GithubConfig, error) {
	token, err := authDeviceFlow()
	if err != nil {
		return nil, err
	}
	cfg := &GithubConfig{
		AccessToken: token,
	}
	return cfg, nil
}
