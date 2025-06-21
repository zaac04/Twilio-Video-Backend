package s3

import (
	"context"
	AppConfig "stargazer/video-recording/config"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GenerateGetPresignedUrl(PresignObj *PresignObject) (url string, err error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AppConfig.App.S3_BUCKET_REGION))

	if err != nil {
		return "", err
	}

	client := s3.NewPresignClient(s3.NewFromConfig(cfg))

	res, err := client.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &PresignObj.Bucket,
		Key:    &PresignObj.Key,
	}, s3.WithPresignExpires(160*time.Hour))

	if err != nil {
		return "", err
	}

	return res.URL, err
}

func GeneratePutPresignedUrl(PresignObj *PresignObject) (url string, err error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AppConfig.App.S3_BUCKET_REGION))

	if err != nil {
		return "", err
	}

	client := s3.NewPresignClient(s3.NewFromConfig(cfg))

	res, err := client.PresignPutObject(context.Background(), &s3.PutObjectInput{
		Bucket: &PresignObj.Bucket,
		Key:    &PresignObj.Key,
	}, s3.WithPresignExpires(160*time.Hour))

	if err != nil {
		return "", err
	}

	return res.URL, err
}
