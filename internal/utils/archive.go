package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExtractTarGz 解压tar.gz文件
func ExtractTarGz(gzipStream io.Reader, destPath string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("创建gzip reader失败: %w", err)
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取tar文件失败: %w", err)
		}

		// 获取文件路径
		path := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
		case tar.TypeReg:
			outFile, err := createFile(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("写入文件失败: %w", err)
			}
			outFile.Close()
		default:
			fmt.Printf("未处理的tar类型: %c in file %s\n", header.Typeflag, path)
		}
	}
	return nil
}

// ExtractTarGzFile 解压tar.gz文件，接受文件路径作为参数
func ExtractTarGzFile(tarGzPath string, destPath string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	return ExtractTarGz(file, destPath)
}

// ExtractZip 解压zip文件
func ExtractZip(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		err := extractZipFile(file, destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// extractZipFile 解压单个zip文件
func extractZipFile(file *zip.File, destPath string) error {
	// 检查文件是否是一个目录
	if file.FileInfo().IsDir() {
		path := filepath.Join(destPath, file.Name)
		return os.MkdirAll(path, 0755)
	}

	// 获取文件路径
	path := filepath.Join(destPath, file.Name)

	// 确保文件的目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 打开源文件
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer dstFile.Close()

	// 复制内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// createFile 创建文件，确保文件的父目录存在
func createFile(path string) (*os.File, error) {
	// 确保目录存在
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	// 创建文件
	return os.Create(path)
}
