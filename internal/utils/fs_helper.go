package utils

import (
	"io"
	"os"
	"path/filepath"
)

// IsDirEntry 判断是否是目录
func IsDirEntry(entry os.DirEntry) bool {
	return entry.IsDir()
}

// IsFileEntry 判断是否是文件
func IsFileEntry(entry os.DirEntry) bool {
	return !entry.IsDir()
}

// CheckDirExists 检查目录是否存在
func CheckDirExists(dirPath string) (bool, error) {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false, err
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CopyDir 复制目录
func CopyDir(src, dst string) error {
	// 创建目标目录
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// 遍历源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
