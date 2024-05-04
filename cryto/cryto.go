package cryto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func Hash(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func New() []byte {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)
	return key
}

func UUID() string {
	buff := make([]byte, 32)
	io.ReadFull(rand.Reader, buff)
	return hex.EncodeToString(buff)
}

func writeStream(blocksize int, stream cipher.Stream, src io.Reader, dst io.Writer) (int, error) {
	var (
		buf = make([]byte, 32*1024)
		written = blocksize
	)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			n, err := dst.Write(buf[:n])
			written += n
			if err != nil {
				return written, err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}
	stream := cipher.NewCTR(block, iv)
	return writeStream(0, stream, src, dst)
}

func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // 16 bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}
	rn, err := dst.Write(iv)
	if err != nil {
		return 0, err
	}
	stream := cipher.NewCTR(block, iv)
	return writeStream(rn, stream, src, dst)
}
