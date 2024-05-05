package store

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jun-hf/distributedstorage/cryto"
)

type KeyPath struct {
	PathName, FileName string
}

func (k KeyPath) FilePath() string {
	return fmt.Sprintf("%s/%s", k.PathName, k.FileName)
}

type StoreOpts struct {
	TransformPathFunc TransformPathFunc
	Root              string
}

type Store struct {
	StoreOpts
}

func New(opts StoreOpts) *Store {
	if len(opts.Root) == 0 {
		opts.Root = defaultRoot
	}
	if opts.TransformPathFunc == nil {
		opts.TransformPathFunc = DefaultPathTransformFunc
	}
	return &Store{StoreOpts: opts}
}

func (s *Store) Path(id string, p KeyPath) string {
	return filepath.Join(s.Root, id, p.PathName)
}

func (s *Store) FilePath(id string, p KeyPath) string {
	return filepath.Join(s.Root, id, p.FilePath())
}

func (s *Store) Has(id, key string) bool {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(id, pathKey)

	_, err := os.Stat(fileP)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) HasPath(path string) bool {
	p := fmt.Sprintf("%s/%s", s.Root, path)
	_, err := os.Stat(p)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Read(id, key string) (io.Reader, error) {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(id, pathKey)
	data, err := os.ReadFile(fileP)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (s *Store) CopyRead(id, key string, dst io.Writer) (int64, error) {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(id, pathKey)

	if !s.Has(id, key) {
		return 0, fmt.Errorf("key %v does not exists", key)
	}
	f, err := os.Open(fileP)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(dst, f)
}

func (s *Store) Delete(id, key string) error {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(id, pathKey)
	return s.deleteFullPath(id, fileP)
}

func (s *Store) ClearAll() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Write(id, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Store) WriteDecrypt(encryptKey []byte, id, key string, r io.Reader) (int64, error) {
	f, err := s.openFileToWrite(id, key)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := cryto.CopyDecrypt(encryptKey, r, f)
	return int64(n), err
}

func (s *Store) writeStream(id, key string, r io.Reader) (int64, error) {
	f, err := s.openFileToWrite(id, key)
	if err != nil {
	  return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func (s *Store) openFileToWrite(id, key string) (*os.File, error) {
	keyPath := s.TransformPathFunc(key)
	path := s.Path(id, keyPath)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}
	fileP := s.FilePath(id, keyPath)
	return os.Create(fileP)
}

func (s *Store) FileSize(id, key string) (int64, error) {
	if !s.Has(id, key) {
		return 0, fmt.Errorf("key %v does not exist", key)
	}
	keyPath := s.TransformPathFunc(key)
	f, err := os.Stat(s.FilePath(id, keyPath))
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

func (s *Store) deleteFullPath(id, fileP string) error {
	if string(fileP[0]) == "/" {
		fileP = fileP[1:]
	}
	stoppingDir := fmt.Sprintf("%v/%+v", s.Root, id)
	for {
		if fileP == stoppingDir {
			return nil
		}
		if _, err := os.Stat(fileP); err != nil {
			return err
		}
		if err := os.RemoveAll(fileP); err != nil {
			return err
		}
		fileP = filepath.Dir(fileP)
	}
}

var (
	defaultRoot              = "storeDir"
	DefaultPathTransformFunc = func(key string) KeyPath {
		return KeyPath{PathName: key, FileName: key}
	}
	SHA1PathTransformFunc = func(key string) KeyPath {
		hash := sha1.Sum([]byte(key))
		hashStr := hex.EncodeToString(hash[:])

		blockSize := 5
		numBlock := len(hashStr) / blockSize
		paths := make([]string, numBlock)

		for i := range numBlock {
			from, to := i*blockSize, i*blockSize+blockSize
			paths[i] = hashStr[from:to]
		}
		return KeyPath{
			PathName: strings.Join(paths, "/"),
			FileName: hashStr,
		}
	}
)

type TransformPathFunc func(string) KeyPath
