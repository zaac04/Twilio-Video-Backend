package auth

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	ServiceName  string
	AuthHost     string
	PrivateKey   *rsa.PrivateKey
	AuthKey      *rsa.PublicKey
	AccessToken  string
	RefreshToken string
	BaseUrl      string
}

func NewService(serviceName, servicePemPath, authPemPath, authHost string, BaseUrl string) (*Service, error) {
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
		BaseUrl:     BaseUrl,
	}

	if err := service.authenticate(); err != nil {
		return nil, err
	}

	return service, nil
}

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

	url := "/login/"

	headers := map[string]string{
		"Content-Type": "application/json",
		"SERVICE":      s.ServiceName,
	}

	response, err := s.makeRequest("POST", url, payload, headers)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	s.AccessToken = response["access_token"].(string)
	s.RefreshToken = response["refresh_token"].(string)
	return nil
}

func (s *Service) signMessage(message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPSS(rand.Reader, s.PrivateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *Service) validateToken(tokenString string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.AuthKey, nil
	})

	if err != nil {
		return nil, err
	}
	return claims, nil
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

func (s *Service) refreshToken() error {
	payload := map[string]string{
		"refresh_token": s.RefreshToken,
	}
	url := "/refresh/"
	headers := map[string]string{"Content-Type": "application/json"}

	response, err := s.makeRequest("POST", url, payload, headers)
	if err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}

	s.AccessToken = response["access_token"].(string)
	s.RefreshToken = response["refresh_token"].(string)
	return nil
}

func (s *Service) Request(method, url string, payload map[string]string) (*http.Response, error) {
	if claims, err := s.validateToken(s.AccessToken); err != nil || claims == nil {
		if claims, err := s.validateToken(s.RefreshToken); err != nil && claims == nil {
			if err := s.authenticate(); err != nil {
				return nil, err
			}
		} else {
			if err := s.refreshToken(); err != nil {
				return nil, err
			}
		}
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + s.AccessToken,
	}
	return s.makeHTTPRequest(method, url, payload, headers)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return key, nil

}

func randomUUID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

func (s *Service) makeRequest(method, endpoint string, payload map[string]string, headers map[string]string) (map[string]interface{}, error) {

	endpoint, err := url.JoinPath(s.BaseUrl, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to json url: %s", err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %s", err)
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func (s *Service) makeHTTPRequest(method, endpoint string, payload map[string]string, headers map[string]string) (*http.Response, error) {

	endpoint, err := url.JoinPath(s.BaseUrl, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to json url: %s", err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %s", err)
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}
