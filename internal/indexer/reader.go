package indexer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"unicode/utf8"
)

type DocReader struct {
	maxFileSize int64
}

func NewDocReader(maxFileSize int64) *DocReader {
	if maxFileSize <= 0 {
		maxFileSize = 1 << 20
	}
	return &DocReader{maxFileSize: maxFileSize}
}

func (r *DocReader) Read(path string) (content string, contentHash string, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("stat file: %w", err)
	}

	readSize := info.Size()
	truncated := false
	if readSize > r.maxFileSize {
		readSize = r.maxFileSize
		truncated = true
	}

	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	buf := make([]byte, readSize)
	if _, err := f.Read(buf); err != nil {
		return "", "", fmt.Errorf("read file: %w", err)
	}

	if truncated {
		extra := make([]byte, 1)
		n, _ := f.Read(extra)
		if n > 0 {
			buf = append(buf, extra[:n]...)
		}
	}

	if !utf8.Valid(buf) {
		content = string(buf)
	} else {
		content = string(buf)
	}

	hash := sha256.Sum256([]byte(content))
	contentHash = fmt.Sprintf("%x", hash)

	return content, contentHash, nil
}
