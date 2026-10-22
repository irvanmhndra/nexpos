// Package storage wraps an S3-compatible object store (Cloudflare R2) used for
// product images and other uploaded assets. Uploads are server-proxied: the API
// validates the bytes then PUTs them to R2; it can also delete objects.
package storage

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client talks to an R2 bucket over the S3 API.
type Client struct {
	s3            *s3.Client
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
	PublicBaseURL string // https://nexpos-cdn.irvanmahendra.com
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
		UsePathStyle: true,
		// R2 doesn't support the SDK's default trailing checksum on PutObject;
		// only send one when an operation explicitly requires it.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
	})

	return &Client{
		s3:            s3Client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}
}

// Upload streams body to key (immutable cache) and returns the object's public URL.
func (c *Client) Upload(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		// Keys are random/immutable, so let the CDN cache forever.
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", err
	}
	return c.PublicURL(key), nil
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

// KeyFromURL is the inverse of PublicURL: it extracts the object key from a URL
// this store owns, or returns "" for an empty/foreign URL (so callers skip it).
func (c *Client) KeyFromURL(url string) string {
	if url == "" || c.publicBaseURL == "" {
		return ""
	}
	prefix := c.publicBaseURL + "/"
	if !strings.HasPrefix(url, prefix) {
		return ""
	}
	return strings.TrimPrefix(url, prefix)
}
