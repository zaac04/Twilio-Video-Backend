package s3

import (
	"context"
	"fmt"
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

type PresignObject struct {
	Bucket string
	Key    string
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

func ParseS3Url(s3url string) (bucket string, key string, base string, err error) {
	parsed, err := url.Parse(s3url)
	if err != nil {
		return
	}
	base = path.Base(parsed.Path) //filename

	switch parsed.Scheme {
	case "s3":
		bucket = parsed.Host
		key = strings.TrimLeft(parsed.Path, "/") //everything after s3://bucket-name/
	case "https":
		hostParts := strings.Split(parsed.Host, ".")
		if len(hostParts) < 4 || hostParts[1] != "s3" {
			return "", "", "", fmt.Errorf("not a valid S3 HTTPS URL")
		}
		bucket = hostParts[0]
		key = strings.TrimLeft(parsed.Path, "/")
	default:
		return "", "", "", fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}

	return bucket, key, base, nil

}
