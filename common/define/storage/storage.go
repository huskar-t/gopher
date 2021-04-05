package storage

import (
	"context"
	"io"
)

type Storage interface {
	PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string, userTags map[string]string) error
	GetObject(ctx context.Context, objectName string) ([]byte, error)
	RemoveObject(ctx context.Context, objectName string) error
	RemoveObjects(ctx context.Context, objectNameChan <-chan string) error
	RemoveByPrefix(ctx context.Context, prefix string) error
}
