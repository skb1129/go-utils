package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Drive struct {
	c *drive.Service
}

func NewDrive() *Drive {
	logger := logs.GetLogger()

	serviceAccount := config.GetMap("gcp.credentials")
	credentialsJSON, err := json.Marshal(serviceAccount)
	if err != nil {
		logger.Fatal("Failed to marshal service account credentials", zap.Error(err))
	}
	options := option.WithAuthCredentialsJSON(option.ServiceAccount, credentialsJSON)
	client, err := drive.NewService(context.TODO(), options)
	if err != nil {
		logger.Fatal("Failed to initialize Drive client", zap.Error(err))
	}

	return &Drive{c: client}
}

func (d *Drive) CreateFolder(ctx context.Context, name string, parentID string) (string, error) {
	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		folder.Parents = []string{parentID}
	}
	f, err := d.c.Files.Create(folder).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return f.Id, nil
}

func (d *Drive) MakePublicReader(ctx context.Context, fileID string) error {
	permission := &drive.Permission{Type: "anyone", Role: "reader"}
	_, err := d.c.Permissions.Create(fileID, permission).Context(ctx).Do()
	return err
}

func (d *Drive) UploadFile(ctx context.Context, name string, parentID string, data io.Reader) (string, error) {
	file := &drive.File{Name: name}
	if parentID != "" {
		file.Parents = []string{parentID}
	}
	f, err := d.c.Files.Create(file).Media(data).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return f.Id, nil
}

func (d *Drive) FindFolderByName(ctx context.Context, name string, parentID string) (string, error) {
	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and trashed=false", name)
	if parentID != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentID)
	}
	r, err := d.c.Files.List().Q(query).Fields("files(id)").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	if len(r.Files) > 0 {
		return r.Files[0].Id, nil
	}
	return "", nil
}
