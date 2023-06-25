package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

type JWTClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

func AuthorizeGoogleJWT(c *fiber.Ctx) (GoogleClaims, error) {
	type GoogleToken struct {
		Credential string `json:"credential"`
	}
	var idtoken GoogleToken
	c.BodyParser(&idtoken)
	google_claims, err := validateGoogleJWT(idtoken.Credential)
	if err != nil {
		return GoogleClaims{}, fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	return google_claims, nil
}

func validateGoogleJWT(token_str string) (GoogleClaims, error) {
	claims_struct := GoogleClaims{}
	token, err := jwt.ParseWithClaims(
		token_str,
		&claims_struct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := getGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return GoogleClaims{}, err
	}
	claims, ok := token.Claims.(*GoogleClaims)
	if !ok {
		return GoogleClaims{}, errors.New("Invalid Google JWT")
	}
	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return GoogleClaims{}, errors.New("iss is invalid")
	}
	if claims.Audience != os.Getenv("GOOGLE_CLIENT_ID") {
		return GoogleClaims{}, errors.New("aud is invalid")
	}
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return GoogleClaims{}, errors.New("JWT is expired")
	}
	return *claims, nil
}

func getGooglePublicKey(keyID string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}
	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	myResp := map[string]string{}
	err = json.Unmarshal(dat, &myResp)
	if err != nil {
		return "", err
	}
	key, ok := myResp[keyID]
	if !ok {
		return "", errors.New("key not found")
	}
	return key, nil
}
