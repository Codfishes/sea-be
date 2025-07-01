package s3

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Interface interface {
	UploadFile(file *multipart.FileHeader, key string) (*UploadResult, error)
	UploadFileFromBytes(data []byte, key, contentType string) (*UploadResult, error)
	UploadFileFromReader(reader io.Reader, key, contentType string, size int64) (*UploadResult, error)
	DownloadFile(key string) ([]byte, error)
	DeleteFile(key string) error
	DeleteFiles(keys []string) error

	GetFileURL(key string) string
	GetPresignedURL(key string, expiration time.Duration) (string, error)
	GetPresignedUploadURL(key string, expiration time.Duration) (string, error)

	ListFiles(prefix string) ([]FileInfo, error)
	FileExists(key string) (bool, error)
	GetFileInfo(key string) (*FileInfo, error)
	CopyFile(sourceKey, destKey string) error
	MoveFile(sourceKey, destKey string) error

	CreateFolder(prefix string) error
	DeleteFolder(prefix string) error
	ListFolders(prefix string) ([]string, error)

	GenerateKey(prefix, filename string) string
	ValidateFileType(filename string, allowedTypes []string) error
	GetFileExtension(filename string) string
	GetContentType(filename string) string
}

type Service struct {
	client   *s3.S3
	uploader *s3manager.Uploader
	config   *Config
}

type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	BucketURL       string
	MaxFileSize     int64
	AllowedTypes    []string
	DefaultACL      string
}

type UploadResult struct {
	Key        string    `json:"key"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	ETag       string    `json:"etag"`
	Location   string    `json:"location"`
	Bucket     string    `json:"bucket"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"content_type"`
	URL          string    `json:"url"`
}

func LoadConfig() *Config {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	bucketURL := os.Getenv("AWS_BUCKET_URL")

	if bucketURL == "" && bucketName != "" {
		bucketURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucketName, region)
	}

	maxFileSize := int64(5 * 1024 * 1024)
	if envSize := os.Getenv("MAX_FILE_SIZE"); envSize != "" {

	}

	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	if envTypes := os.Getenv("ALLOWED_FILE_TYPES"); envTypes != "" {
		allowedTypes = strings.Split(envTypes, ",")
		for i, t := range allowedTypes {
			allowedTypes[i] = strings.TrimSpace(t)
		}
	}

	return &Config{
		Region:          region,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		BucketName:      bucketName,
		BucketURL:       bucketURL,
		MaxFileSize:     maxFileSize,
		AllowedTypes:    allowedTypes,
		DefaultACL:      "public-read",
	}
}

func New() (Interface, error) {
	config := LoadConfig()
	return NewWithConfig(config)
}

func NewWithConfig(config *Config) (Interface, error) {
	if config == nil {
		config = LoadConfig()
	}

	if config.BucketName == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKeyID,
			config.SecretAccessKey,
			"",
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	return &Service{
		client:   client,
		uploader: uploader,
		config:   config,
	}, nil
}

func (s *Service) UploadFile(file *multipart.FileHeader, key string) (*UploadResult, error) {
	if file == nil {
		return nil, fmt.Errorf("file is nil")
	}

	if file.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	if err := s.ValidateFileType(file.Filename, s.config.AllowedTypes); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	if key == "" {
		key = s.GenerateKey("uploads", file.Filename)
	}

	contentType := s.GetContentType(file.Filename)

	return s.UploadFileFromReader(src, key, contentType, file.Size)
}

func (s *Service) UploadFileFromBytes(data []byte, key, contentType string) (*UploadResult, error) {
	reader := bytes.NewReader(data)
	return s.UploadFileFromReader(reader, key, contentType, int64(len(data)))
}

func (s *Service) UploadFileFromReader(reader io.Reader, key, contentType string, size int64) (*UploadResult, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		ACL:         aws.String(s.config.DefaultACL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		Key:        key,
		URL:        s.GetFileURL(key),
		Size:       size,
		ETag:       strings.Trim(*result.ETag, "\""),
		Location:   result.Location,
		Bucket:     s.config.BucketName,
		UploadedAt: time.Now(),
	}, nil
}

func (s *Service) DownloadFile(key string) ([]byte, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	return data, nil
}

func (s *Service) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *Service) DeleteFiles(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	objects := make([]*s3.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = &s3.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	_, err := s.client.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: aws.String(s.config.BucketName),
		Delete: &s3.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}

func (s *Service) GetFileURL(key string) string {
	return fmt.Sprintf("%s/%s", s.config.BucketURL, key)
}

func (s *Service) GetPresignedURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	return req.Presign(expiration)
}

func (s *Service) GetPresignedUploadURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	return req.Presign(expiration)
}

func (s *Service) ListFiles(prefix string) ([]FileInfo, error) {
	result, err := s.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.BucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]FileInfo, len(result.Contents))
	for i, obj := range result.Contents {
		files[i] = FileInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			ETag:         strings.Trim(*obj.ETag, "\""),
			URL:          s.GetFileURL(*obj.Key),
		}
	}

	return files, nil
}

func (s *Service) FileExists(key string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Service) GetFileInfo(key string) (*FileInfo, error) {
	result, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return &FileInfo{
		Key:          key,
		Size:         *result.ContentLength,
		LastModified: *result.LastModified,
		ETag:         strings.Trim(*result.ETag, "\""),
		ContentType:  contentType,
		URL:          s.GetFileURL(key),
	}, nil
}

func (s *Service) CopyFile(sourceKey, destKey string) error {
	_, err := s.client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(s.config.BucketName),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.config.BucketName, sourceKey)),
		Key:        aws.String(destKey),
		ACL:        aws.String(s.config.DefaultACL),
	})
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (s *Service) MoveFile(sourceKey, destKey string) error {

	if err := s.CopyFile(sourceKey, destKey); err != nil {
		return err
	}

	return s.DeleteFile(sourceKey)
}

func (s *Service) CreateFolder(prefix string) error {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(prefix),
		Body:   strings.NewReader(""),
	})
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

func (s *Service) DeleteFolder(prefix string) error {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	files, err := s.ListFiles(prefix)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	keys := make([]string, len(files))
	for i, file := range files {
		keys[i] = file.Key
	}

	return s.DeleteFiles(keys)
}

func (s *Service) ListFolders(prefix string) ([]string, error) {
	result, err := s.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(s.config.BucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	folders := make([]string, len(result.CommonPrefixes))
	for i, cp := range result.CommonPrefixes {
		folders[i] = *cp.Prefix
	}

	return folders, nil
}

func (s *Service) GenerateKey(prefix, filename string) string {
	ext := s.GetFileExtension(filename)
	timestamp := time.Now().Unix()
	uniqueName := fmt.Sprintf("%d_%s", timestamp, filename)

	if ext != "" {

		baseName := strings.TrimSuffix(uniqueName, ext)
		uniqueName = fmt.Sprintf("%s%s", baseName, ext)
	}

	if prefix != "" {
		return fmt.Sprintf("%s/%s", prefix, uniqueName)
	}

	return uniqueName
}

func (s *Service) ValidateFileType(filename string, allowedTypes []string) error {
	if len(allowedTypes) == 0 {
		return nil
	}

	ext := strings.ToLower(s.GetFileExtension(filename))
	contentType := s.GetContentType(filename)

	for _, allowedType := range allowedTypes {
		if strings.Contains(allowedType, "/") {

			if contentType == allowedType {
				return nil
			}
		} else {

			if ext == strings.ToLower(allowedType) || ext == "."+strings.ToLower(allowedType) {
				return nil
			}
		}
	}

	return fmt.Errorf("file type not allowed: %s", ext)
}

func (s *Service) GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

func (s *Service) GetContentType(filename string) string {
	ext := strings.ToLower(s.GetFileExtension(filename))

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

func ExtractKeyFromURL(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fileURL
	}

	return strings.TrimPrefix(parsedURL.Path, "/")
}
