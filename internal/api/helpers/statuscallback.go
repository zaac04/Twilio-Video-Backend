package helpers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/utils"

	"stargazer/video-recording/pkg/aws/mediaconvert"
	"stargazer/video-recording/pkg/twilio"
	"stargazer/video-recording/pkg/yad"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func OnRecordingCompleted(r *http.Request) error {
	participantId := r.FormValue("ParticipantSid")
	recordingType := r.FormValue("Type")

	model := models.ParticipantStatus{}

	if recordingType == "video" {
		model.VideoStatus = true
	} else if recordingType == "audio" {
		model.AudioStatus = true
	}

	//NOTE:Bitmap heap followed by bitmap index scan
	db.ExecuteTransaction(func(tx *gorm.DB) error {
		err := tx.Where("participant_id = ?", participantId).Updates(&model).Error
		return err
	})

	return nil
}

func OnRecordingStarted(r *http.Request) error {

	participantId := r.FormValue("ParticipantSid")
	recordingType := r.FormValue("Type")
	recordingSid := r.FormValue("RecordingSid")
	container := r.FormValue("Container")
	roomName := r.FormValue("RoomName")

	model := models.ParticipantStatus{}
	if recordingType == "video" {
		model.Video = utils.AddExtention(recordingSid, container)
	} else if recordingType == "audio" {
		model.Audio = utils.AddExtention(recordingSid, container)
	}

	//NOTE: query bitmap index scan
	err := db.ExecuteTransaction(func(tx *gorm.DB) error {
		err := tx.Where("participant_id = ?", participantId).Updates(&model).Error
		if err != nil {
			return err
		}

		//NOTE: query  index scan
		err = tx.Where("room_name = ?", roomName).Updates(models.Interview{
			ProcessingStatus: enums.MediaConvertNotStarted,
		}).Error

		return err
	})

	return err
}

func OnParticipantConneted(r *http.Request) error {
	roomName := r.FormValue("RoomName")
	participantId := r.FormValue("ParticipantSid")

	err := db.ExecuteTransaction(func(tx *gorm.DB) error {
		err := tx.Create(&models.ParticipantStatus{
			InterviewRoomName: roomName,
			ParticipantId:     participantId,
		}).Error
		return err
	})

	return err
}

func TriggerMediaConvert(r *http.Request, span *yad.Span) (err error) {
	roomName := r.FormValue("RoomName")
	roomStatus := r.FormValue("RoomStatus")
	roomSid := r.FormValue("RoomSid")
	Event := r.FormValue("StatusCallbackEvent")

	if len(roomStatus) == 0 {
		roomStatus, err = fetchRoomStatusFromTwilio(roomSid)
		if err != nil {
			return fmt.Errorf("failed to fetch room status from Twilio: %v", err)
		}
	}
	if roomStatus != "completed" {
		return fmt.Errorf("room %s is not completed yet, skipping trigger", roomName)
	}

	span.AddEvent("Processing room_closed for room", roomName)

	err = db.ExecuteTransaction(func(d *gorm.DB) error {
		incompleteCount, err := checkIncompleteStatus(d, roomName)
		if err != nil {
			return fmt.Errorf("failed to check incomplete participant status: %v", err)
		}

		if incompleteCount > 0 {
			return fmt.Errorf("triggering failed due to ongoing recording session room %s has %d incomplete participants, skipping trigger", roomName, incompleteCount)
		}
		return nil
	})

	if err != nil {
		return
	}

	err = db.ExecuteTransaction(
		func(d *gorm.DB) error {
			var interview models.Interview

			if err := d.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("room_name = ?", roomName).
				First(&interview).Error; err != nil {

				// Return early if the room doesn't exist
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("room not found: %s", roomName)
				}
				return err
			}

			// If the room's processing status is already set to MediaConvertStarted or is in progress, do not proceed
			if interview.ProcessingStatus != "" && interview.ProcessingStatus != enums.MediaConvertNotStarted {
				return fmt.Errorf("room %s already processing or completed", roomName) // Skip update if already processing or completed
			}

			interview.ProcessingStatus = enums.MediaConvertStarted

			if err := d.Save(&interview).Error; err != nil {
				return fmt.Errorf("error updating room processing status: %v", err)
			}

			span.AddEvent("Room status updated", map[string]string{
				"RoomName": roomName,
				"Status":   enums.MediaConvertStarted,
			})
			return nil
		},
	)

	if err != nil {
		return err
	}

	// Trigger media conversion and update the interview status
	if err := triggerMediaConversion(roomName); err != nil {
		return fmt.Errorf("failed to trigger media conversion: %v", err)
	}

	span.AddEvent("Successfull Trigger", fmt.Sprintf("Successfully triggered media conversion for room: %s during event: %s", roomName, Event))
	return nil
}

func fetchRoomStatusFromTwilio(roomSid string) (string, error) {
	tw := twilio.CreateClient()
	status, err := tw.FetchRoomStatus(roomSid)
	if err != nil {
		log.Printf("Error Fetching Twilio room: %v", err)
		return "", err
	}
	return status, nil
}

func checkIncompleteStatus(tx *gorm.DB, roomName string) (int64, error) {
	var incompleteCount int64
	query := `
		SELECT count(*) AS total
		FROM (
			SELECT 1
			FROM participant_statuses
			WHERE interview_room_name = ? AND (audio_status = ? OR video_status = ?)
			UNION ALL
			SELECT 1
			FROM interviews
			WHERE room_name = ? AND processing_status != ?
		) AS combined;
	`

	err := tx.Raw(query, roomName, false, false, roomName, enums.MediaConvertNotStarted).Scan(&incompleteCount).Error
	return incompleteCount, err
}

func triggerMediaConversion(roomName string) error {

	mc := mediaconvert.CreateClient()

	var participants []models.ParticipantStatus
	var videos []mediaconvert.File
	err := db.ExecuteTransaction(func(tx *gorm.DB) error {
		return tx.Where("interview_room_name = ?", roomName).
			Order("created_at ASC").
			Find(&participants).Error

	})
	if err != nil {
		return err
	}

	for _, participant := range participants {
		video := mediaconvert.File{
			AudioUrl: utils.GetS3Uri(roomName, participant.Audio),
			VideoUrl: utils.GetS3Uri(roomName, participant.Video),
		}
		videos = append(videos, video)
	}
	//NOTE:query ok

	jobId, err := mc.CreateJob(videos, []mediaconvert.Definition{mediaconvert.SD480p, mediaconvert.SD360p}, utils.GetS3SaveUri(roomName))
	if err != nil {
		return err
	}

	return db.ExecuteTransaction(func(tx *gorm.DB) error {
		return tx.Where("room_name = ?", roomName).
			Updates(models.Interview{
				MediaConvertJobId: jobId,
				ProcessingStatus:  enums.MediaConvertOnGoing,
			}).Error
	})
}
