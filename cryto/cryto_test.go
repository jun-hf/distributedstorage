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
	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	res := new(bytes.Buffer)
	_, err = CopyDecrypt(key, dst, res)
	if err != nil {
		t.Fatal(err)
	}
	if res.String() != data {
		t.Fatalf("decryption failed expected (%v) got (%v)", data, res.String())
	}
}