package twilio

import (
	"fmt"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/utils"

	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	"github.com/twilio/twilio-go/client/jwt"
	openapi "github.com/twilio/twilio-go/rest/video/v1"
)

type Twilio struct {
	client *twilio.RestClient
}

func CreateClient() *Twilio {
	accountSid := config.App.TWILIO_ACCOUNT_SID
	apiKey := config.App.TWILIO_API_KEY
	apiSecretKey := config.App.TWILIO_API_SECRET_KEY

	return &Twilio{
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username:   apiKey,
			Password:   apiSecretKey,
			AccountSid: accountSid,
		}),
	}
}

func (tw *Twilio) GenerateAccessToken(room_name string, identity string) (token string, err error) {
	accountSid := config.App.TWILIO_ACCOUNT_SID
	apiKey := config.App.TWILIO_API_KEY
	apiSecretKey := config.App.TWILIO_API_SECRET_KEY

	params := jwt.AccessTokenParams{
		AccountSid:    accountSid,
		SigningKeySid: apiKey,
		Secret:        apiSecretKey,
		Identity:      "user",
	}

	jwtToken := jwt.CreateAccessToken(params)
	videoGrant := &jwt.VideoGrant{
		Room: room_name,
	}

	jwtToken.AddGrant(videoGrant)
	return jwtToken.ToJwt()
}

func (tw *Twilio) FetchRoomStatus(roomSid string) (status string, err error) {
	twRoom, err := tw.client.VideoV1.FetchRoom(roomSid)

	if err != nil || twRoom.Status == nil {
		return "", fmt.Errorf("error fetching room from twilio: %s", err)
	}
	return *twRoom.Status, nil
}

func (tw *Twilio) CreateRoom(roomName string, Expiry int32) (string, error) {
	Room, err := tw.client.VideoV1.CreateRoom(&openapi.CreateRoomParams{
		UniqueName:                  utils.StringPtr(roomName),
		RecordParticipantsOnConnect: utils.BoolPtr(true),
		EmptyRoomTimeout:            utils.IntPtr(60),
		UnusedRoomTimeout:           utils.IntPtr(60),
		MaxParticipants:             utils.IntPtr(1),
		MaxParticipantDuration:      utils.IntPtr(60),
		StatusCallback:              utils.StringPtr(config.App.TWILIO_CALLBACK_URL),
		StatusCallbackMethod:        utils.StringPtr("POST"),
	})

	if err != nil {
		return "", fmt.Errorf("error creating twilio room: %v", err)
	}
	return *Room.Sid, nil
}

func (tw *Twilio) CloseRoom(RoomSid string) error {

	_, err := tw.client.VideoV1.UpdateRoom(RoomSid, &openapi.UpdateRoomParams{
		Status: utils.StringPtr("completed"),
	})

	if err != nil {
		return fmt.Errorf("failed to close twilio room: %v", err)
	}
	return nil
}

func (tw *Twilio) DeleteAllRoom() (roomNames []string, err error) {
	Room, err := tw.client.VideoV1.ListRoom(&openapi.ListRoomParams{})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch twilio rooms: %v", err)
	}

	for _, v := range Room {
		roomNames = append(roomNames, *v.UniqueName)
		_, err := tw.client.VideoV1.UpdateRoom(*v.Sid, &openapi.UpdateRoomParams{
			Status: utils.StringPtr("completed"),
		})
		if err != nil {
			fmt.Println("twilio: failed to update room status: ", err)
			continue
		}
	}
	return roomNames, nil
}

func ValidateRequest(signature string, params map[string]string) (ok bool) {
	validator := client.NewRequestValidator(config.App.TWILIO_API_AUTH_TOKEN)
	return validator.Validate(config.App.TWILIO_CALLBACK_URL, params, signature)
}
