package handlers

import (
	"encoding/json"
	"fmt"
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
	var reqBody schemas.CreateRoom
	db := db.Pg.GetClient()

	err := utils.UnmarshalReqBody(r.Body, &reqBody)
	if err != nil {
		helpers.RespondJsonDecodeErr(err, &ErrMeta)
		return
	}

	res := db.Where("room_name = ?", reqBody.RoomName).Find(&RoomDetails)
	if res.Error != nil {
		helpers.RespondDbFailed(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.DBRetrieveFailed, error_handler.InternalError)
		return
	}

	if len(reqBody.RoomName) == 0 {
		helpers.RespondValidationFailed(fmt.Errorf("room name must not be empty"), &ErrMeta)
		return
	}

	if res.RowsAffected != 0 && RoomDetails.Status != "" {
		helpers.RespondTwilioRoomAlreadyExists(fmt.Errorf("%s", error_handler.TwilioRoomAlreadyExist), &ErrMeta)
		utils.LogError(fmt.Errorf("%s", error_handler.TwilioRoomAlreadyExist), ErrMeta.ReqId, error_handler.TwilioRoomAlreadyExist, error_handler.InternalError)
		return
	}

	//creates a twilio room and respons error if occurs
	token, roomId, err := helpers.CreateRoomHelper(&ErrMeta, &reqBody)
	if err != nil {
		helpers.RespondTwilioRoomCreationError(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.TwilioRoomCreationError, error_handler.InternalError)
		return
	}

	if db.Create(&models.Interview{RoomName: reqBody.RoomName, RoomSid: roomId, Token: token, Status: enums.RoomStatusOnGoing}).Error != nil {
		helpers.RespondDbFailed(err, &ErrMeta)
		utils.LogError(err, ErrMeta.ReqId, error_handler.DBEntryFailed, error_handler.InternalError)
		return
	}

	data, err := json.MarshalIndent(schemas.CreateRoomResp{Token: token, RoomName: reqBody.RoomName}, " ", " ")
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
		helpers.RespondJsonDecodeErr(err, &ErrMeta)
		return
	}

	if len(reqBody.RoomName) == 0 {
		helpers.RespondValidationFailed(fmt.Errorf("room name must not be empty"), &ErrMeta)
		return
	}

	res := db.Where("room_name = ?", reqBody.RoomName).Find(&RoomDetails)
	if res.Error != nil {
		helpers.RespondDbFailed(err, &ErrMeta)
		return
	}

	if res.RowsAffected != 0 && RoomDetails.Status == enums.RoomStatusClosed {
		helpers.RespondTwilioRoomAlreadyClosed(fmt.Errorf("%s", error_handler.TwilioRoomNotFound), &ErrMeta)
		return
	} else if res.RowsAffected == 0 && RoomDetails.Status == "" {
		helpers.RespondTwilioRoomRoomNotFound(fmt.Errorf("%s", error_handler.TwilioRoomNotFound), &ErrMeta)
		return
	}

	err = client.CloseRoom(RoomDetails.RoomSid)

	if err != nil {
		helpers.RespondTwilioRoomCloseError(err, &ErrMeta)
		return
	}

	result := db.Where("room_name = ?", RoomDetails.RoomName).Updates(&models.Interview{
		Status: enums.RoomStatusClosed,
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

func GetRoomDetails(w http.ResponseWriter, r *http.Request) {
	Errmeta := helpers.GenerateErrMeta(r, w)
	room_name := r.URL.Query().Get("room_name")
	fmt.Println(room_name)
	var RoomDetails models.Interview

	if len(room_name) == 0 {
		helpers.RespondQueryParamsNotFound(fmt.Errorf("query params not found"), &Errmeta)
		return
	}

	client := db.Pg.GetClient()

	res := client.Where("room_name = ?", room_name).First(&RoomDetails)

	if res.Error != nil {
		helpers.RespondDbFailed(res.Error, &Errmeta)
		return
	}

	if res.RowsAffected == 0 && RoomDetails.Status == "" {
		helpers.RespondTwilioRoomRoomNotFound(fmt.Errorf("%s", error_handler.TwilioRoomNotFound), &Errmeta)
		return
	}

	data, err := json.Marshal(schemas.RoomDetails{
		RoomName:         room_name,
		Token:            RoomDetails.Token,
		VideoUrl:         RoomDetails.VideoUrl,
		RoomStatus:       RoomDetails.Status,
		ProcessingStatus: RoomDetails.ProcessingStatus,
	})

	if err != nil {
		helpers.RespondJsonEncodeErr(err, &Errmeta)
		return
	}

	helpers.SendResponse(w, data)
}
