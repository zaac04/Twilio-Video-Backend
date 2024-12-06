package utils

import (
	"fmt"
	"log"
	"net/http"
	"stargazer/video-recording/config"
)

func StringPtr(s string) *string                        { return &s }
func IntPtr(i int) *int                                 { return &i }
func BoolPtr(b bool) *bool                              { return &b }
func AddExtention(name string, extention string) string { return name + "." + extention }

func GetS3Uri(roomName, object string) string {
	return fmt.Sprintf("s3://%s/%s/%s/%s", config.App.S3_BUCKET_NAME, config.App.S3_RECORDING_PREFIX, roomName, object)
}

func GetS3SaveUri(roomName string) string {
	return fmt.Sprintf("s3://%s/%s/%s/%s", config.App.S3_BUCKET_NAME, config.App.S3_MEDIA_CONVERT_PREFIX, roomName, "final")
}

func GetClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		if r.Header.Get("X-Forwarded-For") != "" {
			ip = r.Header.Get("X-Forwarded-For")
		} else {
			ip = r.RemoteAddr
		}
	}
	return ip
}

func ExitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
