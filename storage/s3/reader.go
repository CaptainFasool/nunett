package storage

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3ReaderOption func(*S3Reader)

type S3Reader struct {
	client *s3.S3
	bucket string
	key    string
}

type S3ReaderParams struct {
	Bucket string
	Key    string
}

func NewS3Reader(params S3ReaderParams, opts ...S3ReaderOption) *S3Reader {
	reader := &S3Reader{
		bucket: params.Bucket,
		key:    params.Key,
	}

	for _, opt := range opts {
		opt(reader)
	}

	if reader.client == nil {
		reader.client = s3.New(session.Must(session.NewSession()))
	}

	return reader
}

func WithS3Client(client *s3.S3) S3ReaderOption {
	return func(r *S3Reader) {
		r.client = client
	}
}

func (r *S3Reader) Download(outPathVolume string) error {
	outPath, err := os.Create(outPathVolume)
	if err != nil {
		return fmt.Errorf("failed to create outPath: %w", err)
	}
	defer outPath.Close()

	downloader := s3manager.NewDownloaderWithClient(r.client)

	_, err = downloader.Download(outPath, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(r.key),
	})
	if err != nil {
		return fmt.Errorf("failed to download object from S3: %w", err)
	}

	return nil
}

func (r *S3Reader) Size() (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(r.key),
	}

	output, err := r.client.HeadObject(input)
	if err != nil {
		return 0, fmt.Errorf("failed to get object size from S3: %w", err)
	}

	return aws.Int64Value(output.ContentLength), nil
}
