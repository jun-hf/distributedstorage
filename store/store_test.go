package store

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyPath(t *testing.T) {
	keyPath := KeyPath{
		PathName: "path",
		FileName: "testing.go",
	}
	assert.Equal(t, "testing.go", keyPath.FileName)
	assert.Equal(t, "path", keyPath.PathName)
	
	fileP := keyPath.FilePath()
	assert.Equal(t, "path/testing.go", fileP)
}

func TestSHA1PathTransformFunc(t *testing.T) {
	keyPath := SHA1PathTransformFunc("testing.go")
	assert.Equal(t, "8515c/ead95/9aa81/b171e/c2004/ca878/418b0/1b55a", keyPath.PathName)
	assert.Equal(t, "8515cead959aa81b171ec2004ca878418b01b55a", keyPath.FileName)
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		TransformPathFunc: SHA1PathTransformFunc,
		Root: "testStore",
	}
	store := New(opts)
	keyPath := store.TransformPathFunc("hello")
	assert.Equal(t, "testStore/aaf4c/61ddc/c5e8a/2dabe/de0f3/b482c/d9aea/9434d", store.Path(keyPath))
	assert.Equal(t, "testStore/aaf4c/61ddc/c5e8a/2dabe/de0f3/b482c/d9aea/9434d/aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d", store.FilePath(keyPath))

	content := "Inside the file"
	n, err := store.Write("testingFile", strings.NewReader(content))
	if err != nil {
		t.Fatal("Write failed:", err)
	}
	if n != int64(len(content)) {
		t.Fatalf("Write with wrong content expected (%v) got (%v):", int64(len(content)), n)
	}

}