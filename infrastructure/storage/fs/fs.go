package fs

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type FileStorage struct {
	storagePath string
}

func (storage *FileStorage) PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string, userTags map[string]string) (err error) {
	filePath := path.Join(storage.storagePath, objectName)
	if err := os.MkdirAll(path.Dir(filePath), 0666); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return err
	}
	return nil
}

func (storage *FileStorage) GetObject(ctx context.Context, objectName string) ([]byte, error) {
	file, err := os.Open(path.Join(storage.storagePath, objectName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func (storage *FileStorage) RemoveObject(ctx context.Context, objectName string) error {
	return os.Remove(path.Join(storage.storagePath, objectName))
}

func (storage *FileStorage) RemoveObjects(ctx context.Context, objectNameChan <-chan string) error{
	for objectName := range objectNameChan {
		if err := os.Remove(path.Join(storage.storagePath, objectName)); err != nil {
			return err
		}
	}
	return nil
}

func (storage *FileStorage) RemoveByPrefix(ctx context.Context, prefix string) error{
	if err := os.Remove(path.Join(storage.storagePath, prefix)); err != nil {
		return err
	}
	return nil
}
