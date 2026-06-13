// Package storage wraps an S3-compatible object store (Cloudflare R2) used for
// product images and other uploaded assets. Uploads happen client-side via a
// presigned PUT URL; the API only mints the URL and can delete objects.
package storage

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client talks to an R2 bucket over the S3 API.
type Client struct {
	s3            *s3.Client
	presign       *s3.PresignClient
	bucket        string
	publicBaseURL string
}

// Config holds the R2 connection settings (from env).
type Config struct {
	AccountID     string
	AccessKeyID   string
	SecretKey     string
	Bucket        string
	Endpoint      string // https://<account>.r2.cloudflarestorage.com
	PublicBaseURL string // https://cdn.nexpos.irvanmahendra.com
}

// New builds a storage client. Returns nil when not configured (so the rest of
// the app can run without R2 set up yet — the upload endpoint guards on this).
func New(cfg Config) *Client {
	if cfg.Bucket == "" || cfg.Endpoint == "" || cfg.AccessKeyID == "" {
		return nil
	}

	s3Client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(cfg.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, ""),
	})

	return &Client{
		s3:            s3Client,
		presign:       s3.NewPresignClient(s3Client),
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}
}

// PresignPut returns a short-lived URL the client can PUT the object bytes to.
func (c *Client) PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	req, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// Delete removes an object by key (used for orphan cleanup on update/delete).
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// PublicURL builds the CDN URL for an object key.
func (c *Client) PublicURL(key string) string {
	return c.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}
