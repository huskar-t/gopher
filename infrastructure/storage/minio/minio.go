package minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"io/ioutil"
)

type Storage struct {
	client     *minio.Client
	bucketName string
}

func NewStorage(conf *Config, bucketName string) (*Storage, error) {
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, ""),
		Secure: conf.SSL,
	})
	if err != nil {
		return nil, err
	}
	exist, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, err
	}
	if !exist {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}
	return &Storage{client: client, bucketName: bucketName}, nil
}

func (storage *Storage) PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string, userTags map[string]string) (err error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err = storage.client.PutObject(ctx, storage.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType, UserTags: userTags})
	return err
}

func (storage *Storage) GetObject(ctx context.Context, objectName string) ([]byte, error) {
	reader, err := storage.client.GetObject(ctx, storage.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func (storage *Storage) RemoveObject(ctx context.Context, objectName string) error {
	return storage.client.RemoveObject(ctx, storage.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (storage *Storage) RemoveObjects(ctx context.Context, objectNameChan <-chan string) error {
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for objectName := range objectNameChan {
			stat, err := storage.client.StatObject(ctx, storage.bucketName, objectName, minio.StatObjectOptions{})
			if err == nil {
				objectsCh <- stat
			}
		}
	}()

	for err := range storage.client.RemoveObjects(ctx, storage.bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return err.Err
		}
	}
	return nil
}

func (storage *Storage) RemoveByPrefix(ctx context.Context, prefix string) error {
	opts := minio.ListObjectsOptions{
		UseV1:     true,
		Prefix:    prefix,
		Recursive: true,
	}

	objectsCh := storage.client.ListObjects(ctx, storage.bucketName, opts)
	for err := range storage.client.RemoveObjects(ctx, storage.bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return err.Err
		}
	}
	return nil
}
