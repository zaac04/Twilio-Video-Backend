package comms

import (
	"stargazer/video-recording/config"
)

var Auth_SVC *Service
var Interview_SVC *Service
var err error

func AuthenticateService() error {
	Auth_SVC, err = NewService(config.App.SERVICE_NAME, config.App.SERVICE_KEY, config.App.AUTH_KEY, config.App.AUTH_SVC_URL, config.App.AUTH_SVC_URL)
	if err != nil {
		return err
	}

	if err := Auth_SVC.authenticate(); err != nil {
		return err
	}
	Interview_SVC, err = NewService(config.App.INTERVIEW_SVC_NAME, config.App.SERVICE_KEY, config.App.AUTH_KEY, config.App.AUTH_SVC_URL, config.App.INTERVIEW_SVC_URL)
	return err
}

func ValidateForeignRequest(tokenString string) error {
	_, errors := Auth_SVC.validateToken(tokenString)
	return errors
}
