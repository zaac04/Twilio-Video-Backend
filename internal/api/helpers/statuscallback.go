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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func OnRecordingCompleted(r *http.Request) error {
	participantId := r.FormValue("ParticipantSid")
	recordingType := r.FormValue("Type")
	// roomName := r.FormValue("RoomName")

	model := models.ParticipantStatus{}

	if recordingType == "video" {
		model.VideoStatus = true
	} else if recordingType == "audio" {
		model.AudioStatus = true
	}

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

	err := db.ExecuteTransaction(func(tx *gorm.DB) error {
		err := tx.Where("participant_id = ?", participantId).Updates(&model).Error
		if err != nil {
			return err
		}

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

func TriggerMediaConvert(r *http.Request) error {
	roomName := r.FormValue("RoomName")
	roomStatus := r.FormValue("RoomStatus")
	roomSid := r.FormValue("RoomSid")
	Event := r.FormValue("StatusCallbackEvent")

	tx, err := db.Pg.GetTxClient()
	if err != nil {
		return fmt.Errorf("failed to get transaction client: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("panic recovered: %v", r)
		}
	}()

	if len(roomStatus) == 0 {
		roomStatus, err = fetchRoomStatusFromTwilio(roomSid)
		if err != nil {
			return fmt.Errorf("failed to fetch room status from Twilio: %v", err)
		}
	}

	if roomStatus != "completed" {
		log.Printf("Room %s is not completed yet, skipping trigger.", roomName)
		return nil
	}

	log.Println("======================")
	log.Println("Processing room_closed for room:", roomName)

	incompleteCount, err := checkIncompleteStatus(tx, roomName)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check incomplete participant status: %v", err)
	}

	if incompleteCount > 0 {
		log.Printf("Room %s has %d incomplete participants, skipping trigger.", roomName, incompleteCount)
		tx.Rollback()
		return fmt.Errorf("triggering failed due to ongoing recording session")
	}

	err = db.ExecuteTransaction(
		func(tx *gorm.DB) error {
			var interview models.Interview

			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("room_name = ?", roomName).
				First(&interview).Error; err != nil {
				// Return early if the room doesn't exist
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("Room not found: %s", roomName)
					return nil
				}
				return err
			}

			// If the room's processing status is already set to MediaConvertStarted or is in progress, do not proceed
			if interview.ProcessingStatus != "" && interview.ProcessingStatus != enums.MediaConvertNotStarted {
				return fmt.Errorf("room %s already processing or completed", roomName) // Skip update if already processing or completed
			}

			interview.ProcessingStatus = enums.MediaConvertStarted
			if err := tx.Save(&interview).Error; err != nil {

				return fmt.Errorf("error updating room processing status: %v", err)
			}

			log.Printf("Room %s status updated to %s", roomName, enums.MediaConvertStarted)
			return nil
		},
	)

	if err != nil {
		return err
	}

	// Trigger media conversion and update the interview status
	if err := triggerMediaConversion(tx, roomName); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to trigger media conversion: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Successfully triggered media conversion for room: %s during event: %s", roomName, Event)
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
	err := tx.Model(&models.ParticipantStatus{}).
		Joins("JOIN interviews ON interviews.room_name = participant_statuses.interview_room_name").
		Where("(participant_statuses.interview_room_name = ? AND (participant_statuses.audio_status = ? OR participant_statuses.video_status = ?)) OR (interviews.room_name = ? AND interviews.processing_status != ?)",
			roomName,
			false, // Audio incomplete
			false, // Video incomplete
			roomName,
			enums.MediaConvertNotStarted). // Processing not started
		Count(&incompleteCount).Error

	return incompleteCount, err
}

func triggerMediaConversion(tx *gorm.DB, roomName string) error {
	// Start media conversion in the background
	mc := mediaconvert.CreateClient()

	var participants []models.ParticipantStatus
	var videos []mediaconvert.File

	err := tx.Where("interview_room_name = ?", roomName).
		Order("created_at ASC").
		Find(&participants).Error

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

	jobId, err := mc.CreateJob(videos, []mediaconvert.Definition{mediaconvert.SD480p, mediaconvert.SD360p}, utils.GetS3SaveUri(roomName))
	if err != nil {
		return fmt.Errorf("error creating mediaconvert Job")
	}
	return tx.Where("room_name = ?", roomName).
		Updates(models.Interview{
			MediaConvertJobId: jobId,
			ProcessingStatus:  enums.MediaConvertOnGoing,
		}).Error
}
