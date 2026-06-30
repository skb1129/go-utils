package storage

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type Storage struct {
	c *storage.Client
}

func NewStorage() *Storage {
	logger := logs.GetLogger()
	serviceAccount := config.GetMap("gcp.credentials")
	credentialsJSON, err := json.Marshal(serviceAccount)
	if err != nil {
		logger.Fatal("Failed to marshal service account credentials", zap.Error(err))
	}
	options := option.WithAuthCredentialsJSON(option.ServiceAccount, credentialsJSON)
	client, err := storage.NewClient(context.TODO(), options)
	if err != nil {
		logger.Fatal("Failed to initialize GCS client", zap.Error(err))
	}
	return &Storage{c: client}
}

func (s *Storage) Close() error {
	return s.c.Close()
}

func (s *Storage) GetSignedUploadURL(bucket, fileName, contentType string) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(15 * time.Minute),
		ContentType: contentType,
	}
	url, err := s.c.Bucket(bucket).SignedURL(fileName, opts)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *Storage) GetSignedDownloadURL(bucket, fileName string) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	}
	url, err := s.c.Bucket(bucket).SignedURL(fileName, opts)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *Storage) UploadWriter(ctx context.Context, bucket, fileName string) *storage.Writer {
	return s.c.Bucket(bucket).Object(fileName).NewWriter(ctx)
}

func (s *Storage) UploadFile(ctx context.Context, bucket, fileName string, reader io.Reader) error {
	wc := s.UploadWriter(ctx, bucket, fileName)
	if _, err := io.Copy(wc, reader); err != nil {
		wc.Close()
		return err
	}
	return wc.Close()
}

func (s *Storage) DownloadReader(ctx context.Context, bucket, fileName string) (*storage.Reader, error) {
	return s.c.Bucket(bucket).Object(fileName).NewReader(ctx)
}

func (s *Storage) DownloadFile(ctx context.Context, bucket, fileName string) ([]byte, error) {
	rc, err := s.DownloadReader(ctx, bucket, fileName)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (s *Storage) DeleteFile(ctx context.Context, bucket, fileName string) error {
	return s.c.Bucket(bucket).Object(fileName).Delete(ctx)
}
