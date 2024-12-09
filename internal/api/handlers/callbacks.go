package handlers

import (
	"errors"
	"fmt"
	"log"
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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func MediaConvertCallback(w http.ResponseWriter, r *http.Request) {

	var payload schemas.MediaConvertMPDEvent
	err := utils.UnmarshalReqBodyAllowUnknown(r.Body, &payload)

	if err != nil {
		fmt.Println("Error Unmarshalling,", err)
		return
	}

	db.ExecuteTransaction(
		func(tx *gorm.DB) error {
			var interview models.Interview

			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("media_convert_job_id = ?", payload.Detail.JobID).
				First(&interview).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("Job not found: %s", payload.Detail.JobID)
					return nil
				}
				return err
			}

			if (interview.ProcessingStatus == enums.MediaConvertFinished) || (interview.ProcessingStatus == enums.MediaConvertFailed) {
				return fmt.Errorf("media convert job alread completed %s ", payload.Detail.JobID) // Skip update if already processing or completed
			}

			if payload.Detail.Status == "COMPLETE" {
				interview.ProcessingStatus = enums.MediaConvertFinished
				interview.VideoUrl = payload.Detail.OutputGroupDetails[0].PlaylistFilePaths[0]
			} else if payload.Detail.Status == "ERROR" {
				interview.ProcessingStatus = enums.MediaConvertFailed
			}

			if err := tx.Save(&interview).Error; err != nil {

				return fmt.Errorf("error updating room processing status: %v", err)
			}

			log.Printf("Room %s status updated to %s", interview.RoomName, interview.ProcessingStatus)
			return nil
		},
	)

}

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
			utils.LogError(err, ErrMeta.ReqId, error_handler.PgUpdateFailed, error_handler.InternalError)
			return
		}

		err = helpers.TriggerMediaConvert(r)
		if err != nil {
			utils.LogError(err, ErrMeta.ReqId, error_handler.MediaConvertTriggerFailed, error_handler.InternalError)
			return
		}
	}
}
