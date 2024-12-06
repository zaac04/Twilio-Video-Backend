package helpers

import (
	"fmt"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/pkg/twilio"
)

func CreateRoomHelper(ErrMeta *error_handler.ErrorResponseMeta, createRoomInput *schemas.CreateRoom) (token string, RoomSid string, err error) {
	twilio := twilio.CreateClient()

	if createRoomInput.Expiry == 0 {
		createRoomInput.Expiry = 60
	}

	timeout := createRoomInput.Expiry
	expiry := createRoomInput.Expiry * 60 //convert minutes to seconds

	fmt.Println(timeout, expiry)
	RoomSid, err = twilio.CreateRoom(createRoomInput.RoomName, expiry, timeout)
	if err != nil {
		return "", "", err
	}

	token, err = twilio.GenerateAccessToken(createRoomInput.RoomName, "") //TODO:Implement identity if needed
	if err != nil {
		return "", "", err
	}
	return token, RoomSid, nil
}
