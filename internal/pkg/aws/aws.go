package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type AwsResources struct {
	Session *session.Session
}

func (s *AwsResources) SaveToS3(key string, file io.Reader) (string, error) {
	uploader := s3manager.NewUploader(s.Session)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   file,
		Bucket: aws.String("ekanek"),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", err
	}
	return result.Location, nil
}
