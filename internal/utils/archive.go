package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
			Log.Warning(fmt.Sprintf("未处理的tar类型: %c in file %s", header.Typeflag, path))
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

// ExtractExe 处理Windows可执行安装程序
func ExtractExe(exePath, destPath string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 检查是否在Windows系统上
	if runtime.GOOS != "windows" {
		return fmt.Errorf("只能在Windows系统上运行.exe安装程序")
	}

	// 提示用户手动安装
	Log.Warning("\n\n注意：.NET SDK安装程序需要管理员权限才能运行。")
	Log.Warning("请手动运行以下安装程序：")
	Log.Warning(fmt.Sprintf("%s /install /quiet /norestart", exePath))
	Log.Warning("安装完成后，请按任意键继续...")

	// 等待用户按键
	fmt.Scanln()

	// 复制安装程序到目标目录，以便后续使用
	destExePath := filepath.Join(destPath, filepath.Base(exePath))
	if err := CopyFile(exePath, destExePath); err != nil {
		Log.Warning(fmt.Sprintf("复制安装程序到目标目录失败: %v", err))
	}

	// 创建一个标记文件，表示安装已完成
	markerFile := filepath.Join(destPath, "installation_completed.txt")
	if err := os.WriteFile(markerFile, []byte("Installation completed"), 0644); err != nil {
		Log.Warning(fmt.Sprintf("创建标记文件失败: %v", err))
	}

	return nil
}
