package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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

func AuthorizeJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var email string
		email, err := authorizeEmailJWT(c)
		if err == nil {
			log.Println("AUTH: ", email)
			c.Locals("email", email)
			return c.Next()
		}
		email, err = authorizeGoogleJWT(c)
		if err == nil {
			log.Println("AUTH: ", email)
			c.Locals("email", email)
			return c.Next()
		}
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}
}

func authorizeGoogleJWT(c *fiber.Ctx) (string, error) {
	header := c.GetReqHeaders()["Authorization"]
	header_array := strings.Split(header, " ")
	var token string
	if len(header_array) > 1 {
		token = header_array[1]
	} else {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}
	google_claims, err := ValidateGoogleJWT(token)
	email := google_claims.Email
	if err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	return email, nil
}

func authorizeEmailJWT(c *fiber.Ctx) (string, error) {
	header := c.Cookies("hyppo_jwt")
	header_array := strings.Split(header, " ")
	var token string
	if len(header_array) > 1 {
		token = header_array[1]
	} else {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}
	email_claims, err := validateEmailJWT(token)
	if err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	email := email_claims.Email
	return email, nil
}

func GenerateEmailJWT(email string) (string, error) {
	expirationTime := time.Now().Add(24 * 30 * time.Hour)
	claims := &JWTClaims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token_str, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return token_str, nil
}

func validateEmailJWT(signed_token string) (JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		signed_token,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	if err != nil {
		return JWTClaims{}, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		err = errors.New("Invalid token")
		return JWTClaims{}, err
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("Token expired")
		return JWTClaims{}, err
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

func ValidateGoogleJWT(token_str string) (GoogleClaims, error) {
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
