package s3

import (
	"context"
	"net/url"
	"path"
	AppConfig "stargazer/video-recording/config"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/pkg/yad"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type CopyObject struct {
	Bucket       string
	SourcePrefix string
	DestPrefix   string
	Metadata     map[string]string
}

func CopyFiles(CopyObject CopyObject, span *yad.Span) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AppConfig.App.S3_BUCKET_REGION))

	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)
	destPath := CopyObject.DestPrefix

	span.AddEvent("Copying S3 object", map[string]string{
		"Source": CopyObject.SourcePrefix,
		"Dest":   destPath,
	})

	headOutput, err := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(CopyObject.Bucket),
		Key:    aws.String(CopyObject.SourcePrefix),
	})

	if err != nil {
		span.AddEvent(error_handler.S3CopyFailed, err.Error())
		return err
	}

	_, err = client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:            aws.String(CopyObject.Bucket),
		CopySource:        aws.String(CopyObject.Bucket + "/" + CopyObject.SourcePrefix),
		Key:               aws.String(CopyObject.DestPrefix),
		Metadata:          CopyObject.Metadata,
		MetadataDirective: types.MetadataDirectiveReplace,
		ContentType:       headOutput.ContentType,
	})

	if err != nil {
		span.AddEvent(error_handler.S3CopyFailed, err.Error())
		return err
	}

	span.AddEvent("Deleting S3 Object", map[string]string{
		"Source": CopyObject.SourcePrefix,
		"Dest":   CopyObject.Bucket,
	})

	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &CopyObject.Bucket,
		Key:    aws.String(CopyObject.SourcePrefix),
	})

	if err != nil {
		span.AddEvent(error_handler.S3CopyFailed, err.Error())
		return err
	}
	return nil
}

func GetPathsFromUrl(uri string) (sourcePath string, destPath string, err error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	sourcePath = strings.TrimPrefix(parsedURL.Path, "/")
	return sourcePath, path.Base(sourcePath), nil
}
