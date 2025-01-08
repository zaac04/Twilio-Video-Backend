package comms

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var accessToken string
var refreshToken string

type Service struct {
	ServiceName string
	AuthHost    string
	PrivateKey  *rsa.PrivateKey
	AuthKey     *rsa.PublicKey
	BaseUrl     string
}

type AuthResponse struct {
	Data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	} `json:"data"`
}

func NewService(serviceName, servicePemPath, authPemPath, authHost string, baseUrl string) (*Service, error) {
	privateKey, err := loadPrivateKey(servicePemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	authKey, err := loadPublicKey(authPemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth public key: %w", err)
	}

	service := &Service{
		ServiceName: serviceName,
		AuthHost:    authHost,
		PrivateKey:  privateKey,
		AuthKey:     authKey,
		BaseUrl:     baseUrl,
	}

	return service, nil
}

func ExtractToken(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}
	return parts[1], nil
}
