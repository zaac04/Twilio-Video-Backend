package s3

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
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

func CopyFiles(CopyObject CopyObject) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("S3_BUCKET_REGION")))

	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)
	destPath := CopyObject.DestPrefix
	log.Printf("Copying from %s to %s", CopyObject.SourcePrefix, destPath)

	headOutput, err := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(CopyObject.Bucket),
		Key:    aws.String(CopyObject.SourcePrefix),
	})

	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return err
	}

	log.Printf("Deleting key %s from %s", CopyObject.SourcePrefix, CopyObject.Bucket)
	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &CopyObject.Bucket,
		Key:    aws.String(CopyObject.SourcePrefix),
	})

	if err != nil {
		fmt.Println("Error Deleting Object", err)
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
