package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/db"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
)

func GetVideoUrl(w http.ResponseWriter, r *http.Request) {
	ErrMeta := helpers.GenerateSpan(r, w)
	room_name := r.URL.Query().Get("room_name")
	fmt.Println(room_name)
	var video_url string

	ErrMeta.Span.AddEvent("Params", room_name)

	if len(room_name) == 0 {
		helpers.RespondQueryParamsNotFound(fmt.Errorf("query params not found"), &ErrMeta)
		return
	}

	client := db.Pg.GetClient()
	res := client.Model(&models.Interview{}).Where("room_name = ?", room_name).Pluck("video_url", &video_url).Limit(1)

	if res.Error != nil {
		helpers.RespondDbFailed(helpers.GetDBError(res), &ErrMeta)
		return
	}

	if res.RowsAffected == 0 || video_url == "" {
		helpers.RespondTwilioVideoNotFound(fmt.Errorf("%s", error_handler.TwilioRoomNotFound), &ErrMeta)
		return
	}

	video_url = helpers.ConstructVideoUrl(video_url)

	data, err := json.Marshal(schemas.GetVideoUrl{
		VideoUrl: video_url,
	})

	if err != nil {
		helpers.RespondJsonEncodeErr(err, &ErrMeta)
		return
	}

	helpers.SendResponse(w, data)
}
