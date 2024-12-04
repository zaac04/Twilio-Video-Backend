package helpers

import (
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/pkg/twilio"
)

func CreateRoomHelper(ErrMeta *error_handler.ErrorResponseMeta, createRoomInput *schemas.CreateRoom) (token string, RoomSid string, err error) {
	twilio := twilio.CreateClient()
	RoomSid, err = twilio.CreateRoom(createRoomInput.RoomName, createRoomInput.Expiry) //TODO:Implement expiry
	if err != nil {
		return "", "", err
	}

	token, err = twilio.GenerateAccessToken(createRoomInput.RoomName, "") //TODO:Implement identity if needed
	if err != nil {
		return "", "", err
	}
	return token, RoomSid, nil
}
