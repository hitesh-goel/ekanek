package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
)

func SaveToS3(key string, file io.Reader, s *session.Session) (string, error) {
	uploader := s3manager.NewUploader(s)
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

func DownloadFromS3(key string, file *os.File, s *session.Session) error {
	downloader := s3manager.NewDownloader(s)
	_, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String("ekanek"),
		Key:    aws.String(key),
	})
	return err
}
