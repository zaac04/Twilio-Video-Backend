package comms

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"stargazer/video-recording/config"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/utils"
	"time"

	"github.com/golang-jwt/jwt"
)

func (s *Service) authenticate() error {
	message := randomUUID()
	signature, err := s.signMessage(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	payload := map[string]string{
		"data":      message,
		"signature": signature,
	}

	endpoint := "/login/"

	endpoint, err = url.JoinPath(s.AuthHost, endpoint)
	if err != nil {
		return fmt.Errorf("failed to join url: %s", err)
	}
	headers := map[string]string{
		"Content-Type": "application/json",
		"SERVICE":      s.ServiceName,
	}

	response, err := s.makeHTTPRequest("POST", endpoint, payload, headers)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d", response.StatusCode)
	}

	var auth_response AuthResponse
	err = utils.UnmarshalReqBodyAllowUnknown(response.Body, &auth_response)
	if err != nil {
		return fmt.Errorf("%s: %s", error_handler.ErrorDecodingJson, err)
	}

	accessToken = auth_response.Data.AccessToken
	refreshToken = auth_response.Data.RefreshToken

	if config.App.ENV == "DEV" {
		fmt.Println(accessToken)
	}
	return nil
}

func (s *Service) signMessage(message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPSS(rand.Reader, s.PrivateKey, crypto.SHA256, hash[:], &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

func (s *Service) validateToken(tokenString string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.AuthKey, nil
	})
	return claims, err
}

func (s *Service) refreshToken() error {
	payload := map[string]string{
		"refresh_token": refreshToken,
	}

	endpoint := "/refresh/"
	endpoint, err = url.JoinPath(s.AuthHost, endpoint)
	if err != nil {
		return fmt.Errorf("failed to join url: %s", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}

	response, err := s.makeHTTPRequest("POST", endpoint, payload, headers)
	if err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d", response.StatusCode)
	}

	var auth_response AuthResponse
	err = utils.UnmarshalReqBodyAllowUnknown(response.Body, &auth_response)
	if err != nil {
		return fmt.Errorf("%s: %s", error_handler.ErrorDecodingJson, err)
	}

	accessToken = auth_response.Data.AccessToken
	refreshToken = auth_response.Data.RefreshToken
	if config.App.ENV == "DEV" {
		fmt.Println(accessToken)
	}
	return nil
}

func randomUUID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
