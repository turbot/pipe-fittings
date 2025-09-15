package utils

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

func FileModTime(filePath string) (time.Time, error) {
	file, err := os.Stat(filePath)

	if err != nil {
		return time.Time{}, err
	}

	return file.ModTime(), nil
}

// MoveFile moves a file from source to destination.
//
//	It first attempts the movement using OS primitives (os.Rename)
//	If os.Rename fails, it copies the file byte-by-byte to the destination and then removes the source
func MoveFile(source string, destination string) error {
	// try an os.Rename - it is always faster than copy
	err := os.Rename(source, destination)
	if err == nil {
		return nil
	}

	// os.Rename did not work.
	// do a byte-by-byte copy
	srcFile, err := os.OpenFile(source, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	dstFile, err := os.OpenFile(destination, os.O_WRONLY, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return os.Remove(source)
}

func FilenameNoExtension(fileName string) string {
	fileName = path.Base(fileName)
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func AreFilesEqual(file1, file2 string) (bool, error) {
	content1, err := os.ReadFile(file1)
	if err != nil {
		return false, err
	}
	content2, err := os.ReadFile(file2)
	if err != nil {
		return false, err
	}
	return bytes.Equal(content1, content2), nil
}

func EmptyDir(dirPath string) error {
	d, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dirPath, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return &os.PathError{Op: "copy", Path: src, Err: err}
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func CopyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	directory, _ := os.Open(src)
	objects, err := directory.Readdir(-1)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		srcFilePath := filepath.Join(src, obj.Name())
		dstFilePath := filepath.Join(dst, obj.Name())

		if obj.IsDir() {
			err = CopyDir(srcFilePath, dstFilePath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcFilePath, dstFilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func IsValidDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// MoveDirContents moves all entries (files and subdirectories) from srcDir to destDir, preserving structure.
func MoveDirContents(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(srcDir, e.Name())
		destPath := filepath.Join(destDir, e.Name())
		if e.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			// move nested contents recursively
			if err := MoveDirContents(srcPath, destPath); err != nil {
				return err
			}
			// remove empty src dir
			_ = os.Remove(srcPath)
		} else {
			// ensure parent exists
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			if err := os.Rename(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}
