package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"
)

func GetAllRecordings(w http.ResponseWriter, r *http.Request) {
	ErrMeta := helpers.GenerateErrMeta(r, w)

	var RecordingOutput schemas.RecordingOutput
	var RoomInput schemas.RecordingInput
	var total_count int64

	err := utils.UnmarshalReqBody(r.Body, &RoomInput)

	if err != nil {
		helpers.RespondJsonDecodeErr(err, &ErrMeta)
		return
	}

	db := db.Pg.GetClient()
	db.Where("interview_room_name = ?", RoomInput.RoomName).Model(&models.ParticipantStatus{}).Count(&total_count)

	for i := range total_count {
		url := fmt.Sprintf("%s/%s/%d.mpd", "https://cdn.kytt.site", RoomInput.RoomName, i)
		RecordingOutput.URL = append(RecordingOutput.URL, schemas.Cloudfront_url{
			Url: url,
		})
		RecordingOutput.RoomName = RoomInput.RoomName
	}

	data, err := json.Marshal(RecordingOutput)

	if err != nil {
		helpers.RespondJsonEncodeErr(err, &ErrMeta)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
