package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"strings"
	"time"
)

// TODO hardcode as a prod, but have a hidden flags to to override audience and client id
const armoryCliStagingClientId = "sjkd8ufTR3AxHHZz8XZLE0Y8UAIjTM1I"
const armoryAuthScopes = "openid profile email"
const armoryStagingAudience = "https://api.staging.cloud.armory.io"

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Creates a session for Armory Cloud by logging in with your web browser",
	Run: executeLogin,
}

var timeout = 5 * time.Second
var httpClient = http.Client{
	Timeout: timeout,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

type DeviceTokenData struct {
	DeviceCode     			string `json:"device_code"`
	UserCode       			string `json:"user_code"`
	VerificationUri 		string `json:"verification_uri"`
	ExpiresIn       		int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	VerificationUriComplete string `json:"verification_uri_complete"`
}

type AuthErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type AuthSuccessfulResponse struct {
	// AccessToken Encoded JWT / Bearer Token
	AccessToken             string `json:"access_token"`
	// SecondsUtilTokenExpires the number of seconds until the JWT expires, from when it was created by the Auth Server.
	// The JWT has the exact expiration date time
	SecondsUtilTokenExpires int `json:"expires_in"`
}

type ArmoryCloudPrincipalMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"`
	OrgName string `json:"orgName"`
	TokenExpiration time.Time
}

type Jwt struct {
	PrincipalMetadata *ArmoryCloudPrincipalMetadata `json:"https://cloud.armory.io/principal"`
	ExpiresAt int64 `json:"exp"`
}

func executeLogin(cmd *cobra.Command, args []string) {
	deviceTokenResponse := getDeviceCodeFromAuthorizationServer()

	log.Info("You are about to be prompted to verify the following code in your default browser.")
	log.Info("Device Code: " + deviceTokenResponse.UserCode)

	authStartedAt := time.Now()

	// Sleep for 3 seconds so the user has time to read the above message
	time.Sleep(3 * time.Second)

	// Don't pollute our beautiful terminal with garbage
	browser.Stderr = io.Discard
	browser.Stdout = io.Discard
	err := browser.OpenURL(deviceTokenResponse.VerificationUriComplete)
	if err != nil {
		log.Info("Unable to open your default browser, please go to the following URL in a web browser")
		log.Info(deviceTokenResponse.VerificationUriComplete)
	}

	token := pollAuthorizationServerForResponse(deviceTokenResponse, authStartedAt)
	decodeJwtMetadata(token)

}

// getDeviceCodeFromAuthorizationServer retrieves the device token data required to prompt
// the user to go login to Armory Cloud.
func getDeviceCodeFromAuthorizationServer() *DeviceTokenData {
	requestBody, err := json.Marshal(map[string]string{
		"client_id": armoryCliStagingClientId,
		"scope": armoryAuthScopes,
		"audience": armoryStagingAudience,
	})
	if err != nil {
		log.Fatalln("Failed to create request body for Armory authorization server")
	}

	getDeviceCodeRequest, err := http.NewRequest(
		"POST",
		"https://auth.staging.cloud.armory.io/oauth/device/code",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		log.Fatalln("Failed to create request for Armory authorization server")
	}

	getDeviceCodeRequest.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(getDeviceCodeRequest)
	if err != nil {
		log.Fatalln(err)
	}

	dec := json.NewDecoder(resp.Body)
	var deviceTokenResponse DeviceTokenData
	err = dec.Decode(&deviceTokenResponse)
	if err != nil {
		log.Fatalln(err)
	}
	return &deviceTokenResponse
}

func pollAuthorizationServerForResponse(deviceTokenResponse *DeviceTokenData, authStartedAt time.Time) string {
	var secondsAfterAuthStartedAtWhenDeviceFlowExpires = deviceTokenResponse.ExpiresIn * 1000 - 5000
	deviceFlowExpiresTime := authStartedAt.Add(time.Duration(secondsAfterAuthStartedAtWhenDeviceFlowExpires) * time.Second)
	log.Infof("Waiting for user to login")
	for {
		if time.Now().After(deviceFlowExpiresTime) {
			log.Infof("%d", secondsAfterAuthStartedAtWhenDeviceFlowExpires)
			log.Infof(authStartedAt.Local().String())
			log.Infof(deviceFlowExpiresTime.Local().String())
			log.Fatalln("The device flow request has expired.")
		}

		fmt.Print(".")
		time.Sleep(time.Duration(deviceTokenResponse.Interval) * time.Second)

		requestBody, err := json.Marshal(map[string]string{
			"client_id": armoryCliStagingClientId,
			"device_code": deviceTokenResponse.DeviceCode,
			"grant_type": "urn:ietf:params:oauth:grant-type:device_code",
		})
		if err != nil {
			log.Fatalln("Failed to create request body for Armory authorization server")
		}

		getAuthTokenRequest, err := http.NewRequest(
			"POST",
			"https://auth.staging.cloud.armory.io/oauth/token",
			bytes.NewBuffer(requestBody),
		)
		if err != nil {
			log.Fatalln("Failed to create request for Armory authorization server")
		}

		getAuthTokenRequest.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(getAuthTokenRequest)
		if err != nil {
			log.Fatalln(err)
		}

		dec := json.NewDecoder(resp.Body)
		if resp.StatusCode == 200 {
			fmt.Print("\n")
			var authSuccessfulResponse AuthSuccessfulResponse
			err = dec.Decode(&authSuccessfulResponse)
			if err != nil {
				log.Fatalln(err)
			}
			err = resp.Body.Close()
			if err != nil {
				log.Fatalln("failed to close resource")
			}
			return authSuccessfulResponse.AccessToken
		}

		var errorResponse *AuthErrorResponse
		err = dec.Decode(&errorResponse)
		if err != nil {
			log.Fatalln(err)
		}
		err = resp.Body.Close()
		if err != nil {
			log.Fatalln("failed to close resource")
		}

		if errorResponse.Error != "authorization_pending" {
			log.Fatalln("There was an error polling for user auth. Err: %s, Desc: %s", errorResponse.Error, errorResponse.Description)
		}
	}
}

// TODO probably use a JWT library instead of doing it manually
// Stephan was having issues
func decodeJwtMetadata(encodedJwt string) {
	parts := strings.Split(encodedJwt, ".")
	if len(parts) != 3 {
		log.Fatalln("Expected well-formed JWT")
	}
	jwtMeta := parts[1]

	data, err := base64.StdEncoding.DecodeString(jwtMeta)
	if err != nil {
		log.Debug(err)
		log.Fatalln("Failed to decode JWT metadata")
	}

	var jwt Jwt
	dec := json.NewDecoder(bytes.NewReader(data))
	err = dec.Decode(&jwt)
	if err != nil {
		log.Debug(err)
		log.Fatalln("Failed to deserialize principal claim")
	}

	log.Infof("Welcome %s user: %s, your token expires at: %s", jwt.PrincipalMetadata.OrgName, jwt.PrincipalMetadata.Name, time.Unix(jwt.ExpiresAt, 0).Local().String())
}