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

func (s *Store) Path(p KeyPath) string {
	return filepath.Join(s.Root, p.PathName)
}

func (s *Store) FilePath(p KeyPath) string {
	return filepath.Join(s.Root, p.FilePath())
}

func (s *Store) Has(key string) bool {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(pathKey)

	_, err := os.Stat(fileP)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Read(key string) (io.Reader, error) {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(pathKey)
	data, err := os.ReadFile(fileP)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (s *Store) CopyRead(key string, dst io.Writer) (int64, error) {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(pathKey)

	if !s.Has(key) {
		return 0, fmt.Errorf("key %v does not exists", key)
	}
	f, err := os.Open(fileP)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(dst, f)
}

func (s *Store) Delete(key string) error {
	pathKey := s.TransformPathFunc(key)
	fileP := s.FilePath(pathKey)
	return s.deleteFullPath(fileP)
}

func (s *Store) ClearAll() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) WriteDecrypt(encryptKey []byte, key string, r io.Reader) (int64, error) {
	f, err := s.openFileToWrite(key)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := cryto.CopyDecrypt(encryptKey, r, f)
	return int64(n), err
}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	f, err := s.openFileToWrite(key)
	if err != nil {
	  return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func (s *Store) openFileToWrite(key string) (*os.File, error) {
	keyPath := s.TransformPathFunc(key)
	path := s.Path(keyPath)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}
	fileP := s.FilePath(keyPath)
	return os.Create(fileP)
}

func (s *Store) FileSize(key string) (int64, error) {
	if !s.Has(key) {
		return 0, fmt.Errorf("key %v does not exist", key)
	}
	keyPath := s.TransformPathFunc(key)
	f, err := os.Stat(s.FilePath(keyPath))
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

func (s *Store) deleteFullPath(fileP string) error {
	if string(fileP[0]) == "/" {
		fileP = "." + fileP
	}
	for {
		if fileP == s.Root {
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
