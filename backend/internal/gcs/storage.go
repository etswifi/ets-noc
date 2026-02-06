package gcs

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

type Client struct {
	client     *storage.Client
	bucketName string
}

func NewClient(ctx context.Context, bucketName string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// UploadFile uploads a file to GCS
func (c *Client) UploadFile(ctx context.Context, objectName string, reader io.Reader, contentType string) error {
	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType

	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

// GetSignedURL generates a signed URL for downloading a file
func (c *Client) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := c.client.Bucket(c.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// DeleteFile deletes a file from GCS
func (c *Client) DeleteFile(ctx context.Context, objectName string) error {
	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(objectName)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}

// GetFileMetadata retrieves metadata for a file
func (c *Client) GetFileMetadata(ctx context.Context, objectName string) (*storage.ObjectAttrs, error) {
	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(objectName)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return attrs, nil
}
