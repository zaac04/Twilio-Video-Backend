package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("pong"))
}

func MediaConvertCallback(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Unauthorized: missing credentials", http.StatusUnauthorized)
		return
	}

	if username != config.App.MEDIA_CONVERT_BASIC_AUTH_USER || password != config.App.MEDIA_CONVERT_BASIC_AUTH_PASS {
		http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
		return
	}

	var payload schemas.MediaConvertMPDEvent
	err := utils.UnmarshalReqBody(r.Body, &payload)

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
			//TODO: replace url with cloudfront url
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
