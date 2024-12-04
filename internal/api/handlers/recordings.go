package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/pkg/aws/s3"
)

func RecordingStatusCallback(w http.ResponseWriter, r *http.Request) {

	ErrMeta := helpers.GenerateErrMeta(r, w)

	err := r.ParseForm()
	if err != nil {
		utils.LogError(fmt.Errorf("failed to parse form: %v", err), ErrMeta.ReqId, error_handler.FormParseFailed, error_handler.InternalError)
		return
	}

	switch r.FormValue("StatusCallbackEvent") {
	case "participant-connected":
		err = helpers.OnParticipantConneted(r)
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.DBUpdateFailed, error_handler.InternalError)
		}

	case "recording-started":
		err = helpers.OnRecordingStarted(r)
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.DBUpdateFailed, error_handler.InternalError)
		}

	case "room-ended":
		err = helpers.TriggerMediaConvert(r)
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.MediaConvertTriggerFailed, error_handler.InternalError)
			return
		}

	case "recording-completed":
		source, dest, err := s3.GetPathsFromUrl(r.FormValue("MediaExternalLocation"))
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.UrlParseFailed, error_handler.InternalError)
			return
		}

		err = s3.CopyFiles(s3.CopyObject{
			Bucket:       config.App.S3_BUCKET_NAME,
			SourcePrefix: source,
			DestPrefix:   config.App.S3_RECORDING_PREFIX + "/" + r.FormValue("RoomName") + "/" + dest,
			Metadata: map[string]string{
				"room_name":      r.FormValue("RoomName"),
				"timestamp":      r.FormValue("Timestamp"),
				"duration":       r.FormValue("Duration"),
				"participant_id": r.FormValue("ParticipantSid"),
				"room_id":        r.FormValue("RoomSid"),
			},
		})

		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.S3CopyFailed, error_handler.InternalError)
			return
		}

		err = helpers.OnRecordingCompleted(r)
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, enums.PgUpdateFailed, enums.InternalError)
			return
		}

		err = helpers.TriggerMediaConvert(r)
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.MediaConvertTriggerFailed, error_handler.InternalError)
			return
		}
	}
}

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
