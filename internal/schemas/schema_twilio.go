package schemas

type Room struct {
	RoomName string `json:"room_name"`
}

type CloseRoom struct {
	Room
}

type CreateRoom struct {
	Room
	Expiry int32
}

type GetAllRecordings struct {
	RoomSid string `json:"room_sid"`
	Sid     string `json:"room_name"`
}

type CreateRoomResp struct {
	Token    string `json:"jwt_token"`
	RoomName string `json:"room_name"`

	// Info     interface{} `json:"info"`
}

type AllRooms struct {
	Interviews []string `json:"interviews"`
}

type RecordingInput struct {
	Room
}

type RecordingOutput struct {
	URL []Cloudfront_url `json:"urls"`
	Room
}

type Cloudfront_url struct {
	Url string `json:"url"`
}

type MediaConvertMPDEvent struct {
	Detail struct {
		JobID              string `json:"jobId"`  // Required
		Status             string `json:"status"` // Required
		OutputGroupDetails []struct {
			PlaylistFilePaths []string `json:"playlistFilePaths"` // Optional
		} `json:"outputGroupDetails,omitempty"` // Optional
	} `json:"detail"`
}
