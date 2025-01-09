package helpers

import (
	"fmt"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/pkg/twilio"
)

func CreateRoomHelper(ErrMeta *error_handler.ErrorResponseMeta, createRoomInput *schemas.CreateRoom) (token string, RoomSid string, err error) {
	twilio := twilio.CreateClient()
	var timeout int32 = 60
	var expiry int32 = 180 * 60

	if createRoomInput.Expiry > 0 {
		expiry = createRoomInput.Expiry * 60
		timeout = min(createRoomInput.Expiry, timeout)
	}

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
