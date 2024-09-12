package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
)

const (
	FILEDB_DIR_PERMS  = 0755
	FILEDB_FILE_PERMS = 0644
)

type FileDB struct {
	Path        string
	Collections map[string]*FileDBCollection
}

type FileDBCollection struct {
	DB   *FileDB
	Name string
}

func NewFileDB(path string) *FileDB {
	err := os.MkdirAll(path, FILEDB_DIR_PERMS)
	if err != nil {
		log.Fatalf("Could not create FileDB directory %s: %s", path, err.Error())
	}

	collections := make(map[string]*FileDBCollection)

	return &FileDB{
		Path:        path,
		Collections: collections,
	}
}

func (db *FileDB) NewCollection(collectionName string) *FileDBCollection {
	collection := &FileDBCollection{
		DB:   db,
		Name: collectionName,
	}

	path := collection.Folder()
	err := os.MkdirAll(path, FILEDB_DIR_PERMS)
	if err != nil {
		log.Fatalf("Could not create FileDB collection directory %s: %s", path, err.Error())
	}

	db.Collections[collectionName] = collection

	return collection
}

func (c *FileDBCollection) Write(key string, value any) error {
	path := c.pathFor(key)

	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("Could not remove old cache key %s from collection %s: %w", key, c.Name, err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("Could not marshal value for key %s in collection %s: %w", key, c.Name, err)
	}

	err = os.WriteFile(path, data, FILEDB_FILE_PERMS)
	if err != nil {
		return fmt.Errorf("Could not write value for key %s in collection %s: %w", key, c.Name, err)
	}

	return nil
}

func (c *FileDBCollection) Read(key string, val any) (bool, error) {
	path := c.pathFor(key)

	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("Could not read value for key %s in collection %s: %w", key, c.Name, err)
	}

	err = json.Unmarshal(data, val)
	if err != nil {
		return false, fmt.Errorf("Could not unmarshal cached json for key %s in collection %s into %T: %w", key, c.Name, val, err)
	}

	return true, nil
}

func (c *FileDBCollection) List() ([]string, error) {
	entries, err := os.ReadDir(c.Folder())
	if err != nil {
		return nil, fmt.Errorf("Could not list collection %s: %w", c.Name, err)
	}

	keys := make([]string, len(entries))

	for i, entry := range entries {
		keys[i] = strings.TrimSuffix(entry.Name(), ".json")
	}

	return keys, nil
}

func (c *FileDBCollection) Folder() string {
	return fmt.Sprintf("%s/%s", c.DB.Path, c.Name)
}

func (c *FileDBCollection) pathFor(key string) string {
	return fmt.Sprintf("%s/%s.json", c.Folder(), key)
}
