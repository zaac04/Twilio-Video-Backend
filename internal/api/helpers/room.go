package helpers

import (
	"fmt"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/pkg/twilio"
)

func CreateRoomHelper(ErrMeta *error_handler.ErrorResponseMeta, createRoomInput *schemas.CreateRoom) (token string, RoomSid string, err error) {
	twilio := twilio.CreateClient()
	var timeout int32 = 60               //default room time out
	var expiry int32 = 180 * 60          //default 180 minutes
	var minimum_duration int32 = 10 * 60 //minimum duration twilio room needs (https://www.twilio.com/docs/api/errors/53123)

	if createRoomInput.Expiry > 0 {
		expiry = max(minimum_duration, createRoomInput.Expiry*60)
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
