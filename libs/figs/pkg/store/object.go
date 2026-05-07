package store

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thescaffold/gox-apps-figs/pkg/converter"
)

// ObjectProvider stores files in an S3-compatible bucket.
// Configure via environment variables:
//
//	FIGS_S3_BUCKET   — bucket name (required)
//	FIGS_S3_REGION   — AWS region (default: us-east-1)
//	FIGS_S3_ENDPOINT — custom endpoint for S3-compatible stores (e.g. MinIO)
//	AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY — credentials
type ObjectProvider struct{}

func (p *ObjectProvider) Store(payload *Payload, raw *converter.Response) (*Response, error) {
	bucket := os.Getenv("FIGS_S3_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("store: FIGS_S3_BUCKET is not set")
	}
	region := os.Getenv("FIGS_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	endpoint := os.Getenv("FIGS_S3_ENDPOINT")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if accessKey != "" && secretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("store: s3 config: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(cfg, clientOpts...)

	name, _ := payload.Meta["name"].(string)
	key := fmt.Sprintf("%s.%s", name, raw.Extension)

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(raw.Buffer),
		ContentType: aws.String(raw.Mime),
	})
	if err != nil {
		return nil, fmt.Errorf("store: s3 put: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	if endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	}

	return &Response{URL: url, Raw: string(raw.Buffer)}, nil
}
