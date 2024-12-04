package helpers

import (
	"fmt"
	"log"
	"net/http"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/pkg/twilio"

	"gorm.io/gorm"
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

	// Trigger media conversion and update the interview status
	if err := triggerMediaConversion(tx, roomName); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to trigger media conversion: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Successfully triggered media conversion for room: %s", roomName)
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
	//	go mediaconvert.StartConvertions(roomName)

	// Update the interview's processing status
	return tx.Where("room_name = ?", roomName).
		Updates(models.Interview{
			ProcessingStatus: enums.MediaConvertOnGoing,
		}).Error
}
