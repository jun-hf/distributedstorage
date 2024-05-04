package cryto

import (
	"bytes"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	data := "test data"
	src := bytes.NewReader([]byte(data))
	dst := new(bytes.Buffer)
	key := New()
	n, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	if n != 16 + len(data) {
		t.Errorf("invalid byte size")
	}
	res := new(bytes.Buffer)
	n, err = CopyDecrypt(key, dst, res)
	if err != nil {
		t.Fatal(err)
	}
	if n != res.Len() {
		t.Fatalf("invalid written bytes actual (%v), got (%v)", res.Len(), n)
	}
	if res.String() != data {
		t.Fatalf("decryption failed expected (%v) got (%v)", data, res.String())
	}
}
