package files

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
)

type FileManager struct {
	client        *s3.Client
	presignClient *s3.PresignClient
}

var (
	ErrUploadingFile = errors.New("error uploading file to S3")
	ErrGettingFile = errors.New("error getting file link from S3")
)

func MustNewFileManager() *FileManager {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(viper.GetString("s3.region")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), "")),
	)
	if err != nil {
		slog.Error("Error loading default config", "error", err)
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(viper.GetString("s3.endpoint"))
	})

	return &FileManager{
		client:        client,
		presignClient: s3.NewPresignClient(client),
	}
}

// SaveImage saves an image to S3 and returns the URL.
func (f *FileManager) SaveFile(ctx context.Context, file []byte, name string) error {

	_, err := f.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(viper.GetString("s3.bucket")),
		Key:    aws.String(name),
		Body:   bytes.NewReader(file),
	})
	if err != nil {
		slog.Error("Error uploading file to S3", "error", err)
		return errors.Join(ErrUploadingFile, err)
	}

	return nil
}

// GetFileURL returns a presigned URL for the given file.
func (f *FileManager) GetFileURL(ctx context.Context, name string) (string, error) {
	presignedURL, err := f.presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(viper.GetString("s3.bucket")),
		Key:    aws.String(name),
	}, s3.WithPresignExpires(time.Duration(viper.GetInt("s3.file_lifetime"))*time.Minute))
	if err != nil {
		return "", errors.Join(ErrGettingFile, err)
	}

	return presignedURL.URL, nil
}
