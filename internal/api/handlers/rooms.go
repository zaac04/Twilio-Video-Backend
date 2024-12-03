package handlers

import (
	"encoding/json"
	"net/http"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/enums"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/pkg/twilio"
)

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	ErrMeta := helpers.GenerateErrMeta(r, w)
	var RoomDetails models.Interview
	var createRoomInput schemas.CreateRoom
	db := db.Pg.GetClient()

	err := utils.UnmarshalReqBody(r.Body, &createRoomInput)
	if err != nil {
		utils.LogError(err, ErrMeta.ReqId, error_handler.ErrorDecodingJson, error_handler.ErrorDecodingJson)
		return
	}

	res := db.Where("room_name = ?", createRoomInput.RoomName).Find(&RoomDetails)

	if res.Error != nil {
		helpers.RespondDbFailed(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.DBRetrieveFailed, enums.InternalError)
		return
	}

	if res.RowsAffected != 0 && RoomDetails.Status == "" {
		helpers.RespondTwilioRoomAlreadyExists(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.TwilioRoomAlreadyExist, enums.InternalError)
		return
	}

	//creates a twilio room and respons error if occurs
	token, roomId, err := helpers.CreateRoomHelper(&ErrMeta, &createRoomInput)
	if err != nil {
		helpers.RespondTwilioRoomCreationError(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.TwilioRoomCreationError, enums.InternalError)
		return
	}

	err = db.Create(&models.Interview{
		RoomName: createRoomInput.RoomName,
		RoomSid:  roomId,
		Token:    token,
		Status:   enums.RoomStatusOnGoing,
	}).Error

	if err != nil {
		helpers.RespondDbFailed(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.DBEntryFailed, enums.InternalError)
	}

	data, err := json.Marshal(schemas.CreateRoomResp{
		Token:    token,
		RoomName: createRoomInput.RoomName,
	})

	if err != nil {
		helpers.RespondJsonEncodeErr(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.ErrorEncodingJson, error_handler.InternalError)
		return
	}

	helpers.SendResponse(w, data)

}

func CloseRoom(w http.ResponseWriter, r *http.Request) {
	client := twilio.CreateClient()
	db := db.Pg.GetClient()

	ErrMeta := helpers.GenerateErrMeta(r, w)
	var RoomDetails models.Interview
	var reqBody schemas.CloseRoom

	err := utils.UnmarshalReqBody(r.Body, &reqBody)
	if err != nil {
		utils.LogError(err, ErrMeta.ReqId, error_handler.ErrorDecodingJson, error_handler.ErrorDecodingJson)
		return
	}

	err = client.CloseRoom(RoomDetails.RoomSid)

	if err != nil {
		helpers.RespondTwilioRoomCloseError(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.TwilioFail, error_handler.InternalError)
		return
	}

	result := db.Where("room_name = ?", RoomDetails.RoomName).Updates(&models.Interview{
		Status: "closed",
	})

	if result.Error != nil {
		utils.LogError(result.Error, ErrMeta.ReqId, error_handler.DBUpdateFailed, error_handler.InternalError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("closed"))
}

func GetAllRooms(w http.ResponseWriter, r *http.Request) {
	ErrMeta := helpers.GenerateErrMeta(r, w)
	var AllInterviews []models.Interview
	var AllRooms schemas.AllRooms
	db := db.Pg.GetClient()
	db.Where("processing_status = ?", enums.MediaConvertFinished).Find(&AllInterviews).Order("created_at ASC")

	for _, v := range AllInterviews {
		AllRooms.Interviews = append(AllRooms.Interviews, v.RoomName)
	}

	data, err := json.Marshal(AllRooms)
	if err != nil {
		helpers.RespondJsonEncodeErr(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.ErrorEncodingJson, error_handler.InternalError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func DeleteAllRoom(w http.ResponseWriter, r *http.Request) {
	client := twilio.CreateClient()
	ErrMeta := helpers.GenerateErrMeta(r, w)

	rooms, err := client.DeleteAllRoom()
	if err != nil {
		helpers.RespondTwilioRoomCloseError(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.ErrorEncodingJson, error_handler.InternalError)
		return
	}

	db := db.Pg.GetClient()

	for _, v := range rooms {
		result := db.Where("room_sid = ?", v).Updates(&models.Interview{
			Status: "closed",
		})

		if result.Error != nil {
			utils.LogError(result.Error, ErrMeta.ReqId, error_handler.DBUpdateFailed, error_handler.InternalError)
			continue
		}
	}

	data, err := json.MarshalIndent(schemas.GeneralResp{
		Response: enums.Ok,
	}, " ", " ")

	if err != nil {
		helpers.RespondJsonEncodeErr(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.ErrorEncodingJson, error_handler.InternalError)
		return
	}
	helpers.SendResponse(w, data)
}
