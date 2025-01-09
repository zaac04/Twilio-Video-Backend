package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/comms"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/pkg/aws/s3"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// media convert processing callbacks from twilio
func MediaConvertCallback(w http.ResponseWriter, r *http.Request) {
	ErrMeta := helpers.GenerateSpan(r, w)
	defer ErrMeta.Trace.AddTraceToCtx(r)

	var payload schemas.MediaConvertMPDEvent

	if err := utils.UnmarshalReqBodyAllowUnknown(r.Body, &payload); err != nil {
		ErrMeta.Span.AddEvent(error_handler.ErrorDecodingJson, err.Error())
		return
	}

	var interview models.Interview
	db.ExecuteTransaction(
		func(tx *gorm.DB) error {

			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("media_convert_job_id = ?", payload.Detail.JobID).
				First(&interview).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ErrMeta.Span.AddEvent("Job not found", payload.Detail.JobID)
					return nil
				}
				return err
			}

			if (interview.ProcessingStatus == enums.MediaConvertFinished) || (interview.ProcessingStatus == enums.MediaConvertFailed) {
				ErrMeta.Span.AddEvent("media convert job alread completed", payload.Detail.JobID)
				return fmt.Errorf("media convert job alread completed %s ", payload.Detail.JobID) // Skip update if already processing or completed
			}

			if payload.Detail.Status == "COMPLETE" {
				interview.ProcessingStatus = enums.MediaConvertFinished
				interview.VideoUrl = payload.Detail.OutputGroupDetails[0].PlaylistFilePaths[0]
			} else if payload.Detail.Status == "ERROR" {
				interview.ProcessingStatus = enums.MediaConvertFailed
			}

			if err := tx.Save(&interview).Error; err != nil {
				ErrMeta.Span.AddEvent("error updating room processing status", err.Error())
				return fmt.Errorf("error updating room processing status: %v", err)
			}
			ErrMeta.Span.AddEvent("Room status updated", interview.ProcessingStatus)
			return nil
		},
	)

	payloads := map[string]string{
		"recording_url": helpers.ConstructVideoUrl(interview.VideoUrl),
	}

	resp, err := comms.Interview_SVC.Request("POST", fmt.Sprintf("/api/v1/interviews/video/%s/", interview.RoomName), payloads)

	if err != nil {
		ErrMeta.Span.AddEvent("error in making interservice call", err.Error())
		return
	}

	body, err := utils.DecodeReqBodyAsString(resp.Body)
	if err != nil {
		ErrMeta.Span.AddEvent("error decoding response", err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		ErrMeta.Span.AddEvent("error in making interservice call", map[string]string{
			"StatusCode": strconv.Itoa(resp.StatusCode),
			"Body":       body,
		})
		return
	}

	ErrMeta.Span.AddEvent("Successfully made request with payload", map[string]string{
		"response": body,
	})
}

// Recording Callbacks from twilio
func RecordingStatusCallback(w http.ResponseWriter, r *http.Request) {

	ErrMeta := helpers.GenerateSpan(r, w)
	defer ErrMeta.Trace.AddTraceToCtx(r)

	err := r.ParseForm()
	if err != nil {
		ErrMeta.Span.AddEvent(error_handler.FormParseFailed, err.Error())
		return
	}
	ErrMeta.Span.AddEvent("Event", r.FormValue("StatusCallbackEvent"))

	switch r.FormValue("StatusCallbackEvent") {
	case "participant-connected":
		err = helpers.OnParticipantConneted(r)
		if err != nil {
			ErrMeta.Span.AddEvent(error_handler.DBUpdateFailed, err.Error())
			return
		}

	case "recording-started":
		err = helpers.OnRecordingStarted(r)
		if err != nil {
			ErrMeta.Span.AddEvent(error_handler.DBUpdateFailed, err.Error())
			return
		}

	case "room-ended":
		err = helpers.TriggerMediaConvert(r, ErrMeta.Span)
		if err != nil {
			ErrMeta.Span.AddEvent(error_handler.MediaConvertTriggerFailed, err.Error())
			return
		}

	case "recording-completed":
		source, dest, err := s3.GetPathsFromUrl(r.FormValue("MediaExternalLocation"))
		if err != nil {
			ErrMeta.Span.AddEvent(error_handler.UrlParseFailed, err.Error())
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
			ErrMeta.Span.AddEvent(error_handler.S3CopyFailed, err.Error())
			return
		}

		err = helpers.OnRecordingCompleted(r)
		if err != nil {
			ErrMeta.Span.AddEvent(error_handler.PgUpdateFailed, err.Error())
			return
		}

		err = helpers.TriggerMediaConvert(r, ErrMeta.Span)
		if err != nil {
			ErrMeta.Span.AddEvent(error_handler.MediaConvertTriggerFailed, err.Error())
			return
		}
	}
}
