package comms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	error_handler "stargazer/video-recording/internal/error"
	"time"
)

func (s *Service) Request(method, endpoint string, payload map[string]string) (*http.Response, error) {
	if claims, err := s.validateToken(accessToken); err != nil || claims == nil {
		if claims, err := s.validateToken(refreshToken); err != nil && claims == nil {
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
		"Authorization": "Bearer " + accessToken,
	}

	endpoint, err = url.JoinPath(s.BaseUrl, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to join url: %s", err)
	}

	return s.makeHTTPRequest(method, endpoint, payload, headers)
}

func (s *Service) makeHTTPRequest(method, endpoint string, payload map[string]string, headers map[string]string) (*http.Response, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", error_handler.ErrorEncodingJson, err)
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
