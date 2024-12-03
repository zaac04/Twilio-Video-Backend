package structs

type ReqContext struct {
	ReqID     string
	UserId    string
	SessionId string
	Role      string
}

type ReqCtxKey string

type RecordingCallback struct {
	RoomName      string `json:"room_name"`
	Duration      string `json:"duration"`
	ParticipantId string `json:"participant_id"`
	RoomId        string `json:"room_id"`
	Timestamp     string `json:"time_stamp"`
	Url           string `json:"url"`
}
