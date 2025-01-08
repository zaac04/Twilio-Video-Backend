package helpers

import (
	"net/url"
	"stargazer/video-recording/config"
	"strings"
)

func ConstructVideoUrl(video_url string) string {
	ur, _ := url.Parse(video_url)
	path := strings.TrimPrefix(ur.Path, "/"+config.App.S3_MEDIA_CONVERT_PREFIX)
	video_url, _ = url.JoinPath(config.App.CLOUDFRONT_RESOURCE_URL, path)
	return video_url
}
