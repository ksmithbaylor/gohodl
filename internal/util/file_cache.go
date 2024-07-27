package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const CACHE_DATA_PATH = ".cache"
const CACHE_DIR_PERMS = 0755
const CACHE_FILE_PERMS = 0644

type FileCache struct {
	CollectionName string
}

func NewFileCache(collectionName string) (*FileCache, error) {
	fc := FileCache{
		CollectionName: collectionName,
	}

	err := os.MkdirAll(fc.folder(), CACHE_DIR_PERMS)
	if err != nil {
		return nil, fmt.Errorf("Could not create cache directory %s: %w", CACHE_DATA_PATH, err)
	}

	return &fc, nil
}

func (fc *FileCache) Write(key string, value any) error {
	path := fc.pathFor(key)

	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("Could not remove old cache key %s from collection %s: %w", key, fc.CollectionName, err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("Could not marshal value for key %s in collection %s: %w", key, fc.CollectionName, err)
	}

	err = os.WriteFile(path, data, CACHE_FILE_PERMS)
	if err != nil {
		return fmt.Errorf("Could not write value for key %s in collection %s: %w", key, fc.CollectionName, err)
	}

	return nil
}

func (fc *FileCache) Read(key string, val any) (bool, error) {
	path := fc.pathFor(key)

	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("Could not read value for key %s in collection %s: %w", key, fc.CollectionName, err)
	}

	err = json.Unmarshal(data, val)
	if err != nil {
		return false, fmt.Errorf("Could not unmarshal cached json for key %s in collection %s into %T: %w", key, fc.CollectionName, val, err)
	}

	return true, nil
}

func (fc *FileCache) folder() string {
	return fmt.Sprintf("%s/%s", CACHE_DATA_PATH, fc.CollectionName)
}

func (fc *FileCache) pathFor(key string) string {
	return fmt.Sprintf("%s/%s.json", fc.folder(), key)
}
