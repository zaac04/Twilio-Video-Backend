package helpers

import (
	"fmt"
	"stargazer/video-recording/internal/structs"
	"stargazer/video-recording/pkg/aws/s3"
)

func GeneratePresignedUrls(s3_url string, room_name string) (urls structs.AnalyzeVideoPresignedUrls, err error) {
	bucket, key, _, err := s3.ParseS3Url(s3_url)

	if err != nil {
		return
	}

	urls.Source_Url, err = s3.GenerateGetPresignedUrl(&s3.PresignObject{
		Bucket: bucket,
		Key:    key,
	})

	if err != nil {
		return
	}

	urls.Destination_url, err = s3.GeneratePutPresignedUrl(&s3.PresignObject{
		Bucket: bucket,
		Key:    fmt.Sprintf("analyzed-data/%s.json", room_name),
	})

	if err != nil {
		return
	}

	return urls, err
}
