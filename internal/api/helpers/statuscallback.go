package helpers

import (
	"fmt"
	"net/http"
	"stargazer/video-recording/internal/enums"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/utils"

	"gorm.io/gorm"
)

func OnRecordingCompleted(r *http.Request, tx *gorm.DB) error {
	participantId := r.FormValue("ParticipantSid")
	recordingType := r.FormValue("Type")
	// roomName := r.FormValue("RoomName")

	model := models.ParticipantStatus{}

	if recordingType == "video" {
		model.VideoStatus = true
	} else if recordingType == "audio" {
		model.AudioStatus = true
	}

	err := tx.Where("participant_id = ?", participantId).Updates(&model).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error during transaction: %v", err)
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func TriggerMediaConvert(r *http.Request) error {
	roomName := r.FormValue("RoomName")
	status := r.FormValue("RoomStatus")

	ok, err := CheckTriggerEligibility(status, roomName, r.FormValue("RoomSid"))

	if err != nil {
		return err
	}

	if ok {
		HelperTriggerMediaConvert(roomName)
	}

	return nil
}

func OnRecordingStarted(r *http.Request, tx *gorm.DB) error {

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

	err := tx.Where("participant_id = ?", participantId).Updates(&model).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error during transaction: %v", err)
	}

	err = tx.Where("room_name = ?", roomName).Updates(models.Interview{
		ProcessingStatus: enums.MediaConvertNotStarted,
	}).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error during transaction: %v", err)
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return err
}

func OnParticipantConneted(r *http.Request, tx *gorm.DB) error {
	roomName := r.FormValue("RoomName")
	participantId := r.FormValue("ParticipantSid")

	err := tx.Create(&models.ParticipantStatus{
		InterviewRoomName: roomName,
		ParticipantId:     participantId,
	}).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error during transaction: %v", err)
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
