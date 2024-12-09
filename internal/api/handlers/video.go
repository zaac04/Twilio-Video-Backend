package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/api/helpers"
	"stargazer/video-recording/internal/db"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/schemas"
	"strings"
)

func GetVideoUrl(w http.ResponseWriter, r *http.Request) {
	Errmeta := helpers.GenerateErrMeta(r, w)
	room_name := r.URL.Query().Get("room_name")
	fmt.Println(room_name)
	var video_url string

	if len(room_name) == 0 {
		helpers.RespondQueryParamsNotFound(fmt.Errorf("query params not found"), &Errmeta)
		return
	}

	client := db.Pg.GetClient()
	res := client.Model(&models.Interview{}).Where("room_name = ?", room_name).Pluck("video_url", &video_url).Limit(1)

	if res.Error != nil {
		helpers.RespondDbFailed(res.Error, &Errmeta)
		return
	}

	if res.RowsAffected == 0 || video_url == "" {
		helpers.RespondTwilioVideoNotFound(fmt.Errorf("%s", error_handler.TwilioRoomNotFound), &Errmeta)
		return
	}

	ur, _ := url.Parse(video_url)
	path := strings.TrimPrefix(ur.Path, "/"+config.App.S3_MEDIA_CONVERT_PREFIX)
	video_url, _ = url.JoinPath(config.App.CLOUDFRONT_RESOURCE_URL, path)

	data, err := json.Marshal(schemas.GetVideoUrl{
		VideoUrl: video_url,
	})

	if err != nil {
		helpers.RespondJsonEncodeErr(err, &Errmeta)
		return
	}

	helpers.SendResponse(w, data)
}
