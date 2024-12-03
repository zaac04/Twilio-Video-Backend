package helpers

import (
	"fmt"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/pkg/aws/mediaconvert"
	"stargazer/video-recording/pkg/twilio"

	"gorm.io/gorm"
)

func CreateRoomHelper(ErrMeta *error_handler.ErrorResponseMeta, createRoomInput *schemas.CreateRoom) (token string, RoomSid string, err error) {
	twilio := twilio.CreateClient()
	RoomSid, err = twilio.CreateRoom(createRoomInput.RoomName, createRoomInput.Expiry) //TODO:Implement expiry
	if err != nil {
		return "", "", err
	}

	token, err = twilio.GenerateAccessToken(createRoomInput.RoomName, "") //TODO:Implement identity if needed
	if err != nil {
		return "", "", err
	}
	return token, RoomSid, nil
}

func HelperTriggerMediaConvert(roomName string) (err error) {

	mc := mediaconvert.CreateClient()
	fmt.Println("mediaconvert triggered", roomName)

	videos, err := GetMediaConvertFiles(roomName)
	if err != nil {
		return err
	}

	jobId, err := mc.CreateJob(videos, []mediaconvert.Definition{mediaconvert.SD480p, mediaconvert.SD360p}, utils.GetS3SaveUri(roomName))
	if err != nil {
		return fmt.Errorf("error creating mediaconvert Job")
	}

	fmt.Println(jobId)

	err = db.ExecuteTransaction(func(d *gorm.DB) error {
		return d.Where("room_name = ?", roomName).Updates(models.Interview{
			MediaConvertJobId: jobId,
		}).Error
	})

	return err
}

func GetMediaConvertFiles(roomName string) (videos []mediaconvert.File, err error) {
	var participants []models.ParticipantStatus

	err = db.ExecuteTransaction(func(d *gorm.DB) error {
		err = d.Where("interview_room_name = ?", roomName).
			Order("created_at ASC").
			Find(&participants).Error
		return err
	})

	if err != nil {
		return
	}

	for _, participant := range participants {
		video := mediaconvert.File{
			AudioUrl: utils.GetS3Uri(roomName, participant.Audio),
			VideoUrl: utils.GetS3Uri(roomName, participant.Video),
		}
		videos = append(videos, video)
	}
	return
}

func CheckTriggerEligibility(status, roomName, roomSid string) (bool, error) {
	var err error
	var incompleteCount int64

	tw := twilio.CreateClient()
	if len(status) == 0 {
		status, err = tw.FetchRoomStatus(roomSid)
		if err != nil {
			return false, err
		}
	}

	if status == "completed" {

		err := db.ExecuteTransaction(func(d *gorm.DB) error {
			err = d.Model(&models.ParticipantStatus{}).
				Joins("JOIN interviews ON interviews.room_name = participant_statuses.interview_room_name").
				Where("(participant_statuses.interview_room_name = ? AND (participant_statuses.audio_status = ? OR participant_statuses.video_status = ?)) OR (interviews.room_name = ? AND interviews.processing_status != ?)",
					roomName,
					false,
					false,
					roomName,
					enums.MediaConvertNotStarted).
				Count(&incompleteCount).Error

			if incompleteCount == 0 {
				err = d.Where("room_name = ?", roomName).Updates(models.Interview{
					ProcessingStatus: enums.MediaConvertOnGoing,
				}).Error
			}
			return err
		})

		if err != nil {
			return false, err
		}

		fmt.Println("incomplete count", incompleteCount)

		if incompleteCount != 0 {
			return false, fmt.Errorf("triggering failed due to ongoing recording session")
		}
	}
	return true, nil
}
