package ociinstaller

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const defaultMaxDecompressedSize = 500 << 20 // 500 MB size limit for decompressed data

// define option pattern
type gzipConfig struct {
	maxDecompressedSize int64
}
type UngzipOption func(*gzipConfig)

func WithMaxDecompressedSize(size int64) UngzipOption {
	return func(c *gzipConfig) {
		c.maxDecompressedSize = size
	}
}

func Ungzip(sourceFile string, destDir string, opts ...UngzipOption) (string, error) {
	cfg := &gzipConfig{
		maxDecompressedSize: defaultMaxDecompressedSize,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	r, err := os.Open(sourceFile)
	if err != nil {
		return "", err
	}
	defer r.Close()

	uncompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer uncompressedStream.Close()

	// Limit the amount of data being written to prevent decompression bombs
	limitedReader := io.LimitReader(uncompressedStream, cfg.maxDecompressedSize)

	destFile := filepath.Join(destDir, uncompressedStream.Name)
	outFile, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	n, err := io.Copy(outFile, limitedReader)
	if err != nil {
		return "", err
	}
	// Check if the data was truncated due to the limit
	if n == cfg.maxDecompressedSize {
		return "", fmt.Errorf("data was truncated: limit of %d bytes reached", cfg.maxDecompressedSize)
	}

	return destFile, nil
}

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// MoveFileWithinPartition moves a file within an fs partition. panics if movement is attempted between partitions
// this is done separately to achieve performance benefits of os.Rename over reading and writing content
func MoveFileWithinPartition(sourcePath, destPath string) error {
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("error moving file: %s", err)
	}
	return nil
}

// MoveFolderWithinPartition moves a folder within an fs partition. panics if movement is attempted between partitions
// this is done separately to achieve performance benefits of os.Rename over reading and writing content
func MoveFolderWithinPartition(sourcePath, destPath string) error {
	sourceinfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(destPath, sourceinfo.Mode()); err != nil {
		return err
	}

	directory, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source dir: %s", err)
	}
	directory.Close()

	defer os.RemoveAll(sourcePath)

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		relPath, _ := filepath.Rel(sourcePath, path)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(destPath, relPath), info.Mode())
		}
		return MoveFileWithinPartition(filepath.Join(sourcePath, relPath), filepath.Join(destPath, relPath))
	})
}
