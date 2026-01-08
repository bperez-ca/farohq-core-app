package s3

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"farohq-core-app/internal/domains/files/domain/ports/outbound"
)

// Storage implements the outbound.Storage interface using AWS S3
type Storage struct {
	client *s3.Client
}

// NewStorage creates a new S3 storage adapter
func NewStorage(region, bucket string) (outbound.Storage, error) {
	// Check for local development settings
	awsEndpoint := os.Getenv("AWS_ENDPOINT_URL")

	var awsCfg aws.Config
	var err error

	// For local development with LocalStack or similar
	if awsEndpoint != "" {
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL: awsEndpoint,
					}, nil
				})),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
				},
			}),
		)
		if err != nil {
			return nil, err
		}
	} else {
		// For production AWS
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
		if err != nil {
			return nil, err
		}
	}

	client := s3.NewFromConfig(awsCfg)

	return &Storage{
		client: client,
	}, nil
}

// GeneratePresignedURL generates a pre-signed URL for uploading a file
func (s *Storage) GeneratePresignedURL(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, map[string]string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", nil, err
	}

	// Convert headers to map
	headers := make(map[string]string)
	for key, values := range request.SignedHeader {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return request.URL, headers, nil
}

// DeleteFile deletes a file from storage
func (s *Storage) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var notFound *types.NoSuchKey
		if err != nil && notFound != nil {
			return err
		}
		return err
	}

	return nil
}

