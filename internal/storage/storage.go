package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"webshop/internal/config"
)

type Storage interface {
	Upload(ctx context.Context, filename string, r io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, fileURL string) error
}

type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

type LocalStorage struct {
	uploadDir string
	baseURL   string
}

func NewStorage(cfg *config.Config) (Storage, error) {
	// Gebruik Cloudflare R2 indien geconfigureerd
	if cfg.R2AccountID != "" && cfg.R2AccessKeyID != "" && cfg.R2SecretAccessKey != "" && cfg.R2BucketName != "" {
		endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)

		customResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               endpoint,
					SigningRegion:     "auto",
					HostnameImmutable: true,
				}, nil
			},
		)

		sdkConfig, err := awsConfig.LoadDefaultConfig(
			context.TODO(),
			awsConfig.WithEndpointResolverWithOptions(customResolver),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.R2AccessKeyID,
				cfg.R2SecretAccessKey,
				"",
			)),
			awsConfig.WithRegion("auto"),
		)
		if err != nil {
			return nil, fmt.Errorf("fout bij opzetten R2 client: %w", err)
		}

		client := s3.NewFromConfig(sdkConfig)
		publicURL := strings.TrimSuffix(cfg.R2PublicURL, "/")
		return &R2Storage{
			client:    client,
			bucket:    cfg.R2BucketName,
			publicURL: publicURL,
		}, nil
	}

	// Lokale fallback map
	localDir := "./uploads"
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return nil, fmt.Errorf("kan lokale uploads map niet aanmaken: %w", err)
	}

	return &LocalStorage{
		uploadDir: localDir,
		baseURL:   cfg.BaseURL,
	}, nil
}

func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	randomBytes := make([]byte, 8)
	_, _ = rand.Read(randomBytes)
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s%s", timestamp, hex.EncodeToString(randomBytes), ext)
}

// R2 Implementatie
func (r *R2Storage) Upload(ctx context.Context, originalFilename string, reader io.Reader, contentType string) (string, error) {
	filename := generateUniqueFilename(originalFilename)
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return "", err
	}

	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload naar R2 mislukt: %w", err)
	}

	if r.publicURL != "" {
		return fmt.Sprintf("%s/%s", r.publicURL, filename), nil
	}
	return filename, nil
}

func (r *R2Storage) Delete(ctx context.Context, fileURL string) error {
	key := filepath.Base(fileURL)
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}

// Lokale Implementatie
func (l *LocalStorage) Upload(ctx context.Context, originalFilename string, reader io.Reader, contentType string) (string, error) {
	filename := generateUniqueFilename(originalFilename)
	targetPath := filepath.Join(l.uploadDir, filename)

	out, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("bestand kon lokaal niet aangemaakt worden: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", fmt.Errorf("fout bij wegschrijven lokaal bestand: %w", err)
	}

	return fmt.Sprintf("%s/uploads/%s", strings.TrimSuffix(l.baseURL, "/"), filename), nil
}

func (l *LocalStorage) Delete(ctx context.Context, fileURL string) error {
	filename := filepath.Base(fileURL)
	targetPath := filepath.Join(l.uploadDir, filename)
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}