package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"svm/internal/config"
	"svm/internal/utils"
)

// JavaVersion 表示Java版本信息
type JavaVersion struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Download string `json:"download"`
}

// JavaSDKProvider 实现了SDKProvider接口
type JavaSDKProvider struct {
	config *config.Config
}

// NewJavaSDK 创建一个新的Java SDK
func NewJavaSDK() SDK {
	provider := &JavaSDKProvider{
		config: nil, // 这里为空，会由BaseSDK初始化时设置
	}

	return &javaSDK{
		BaseSDK: *NewBaseSDK("java", provider, DefaultVersionPrefixHandlers()),
	}
}

// javaSDK 是Java SDK的具体实现
type javaSDK struct {
	BaseSDK
}

// GetVersionList 实现SDKProvider接口，获取所有可用的Java版本
func (p *JavaSDKProvider) GetVersionList() ([]string, error) {
	// 从AdoptOpenJDK API获取版本列表
	url := "https://api.adoptium.net/v3/info/available_releases"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var data struct {
		AvailableReleases []int `json:"available_releases"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}

	var versions []string
	for _, v := range data.AvailableReleases {
		versions = append(versions, fmt.Sprintf("%d", v))
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versions)

	return versions, nil
}

// GetAllVersionList 实现SDKProvider接口，获取所有可用的Java版本（不过滤）
func (p *JavaSDKProvider) GetAllVersionList() ([]string, error) {
	// 对于Java，GetVersionList已经返回所有版本，不需要额外过滤
	// 这里直接调用GetVersionList
	return p.GetVersionList()
}

// GetDownloadURL 构建Java下载URL
func (p *JavaSDKProvider) GetDownloadURL(version, osName, arch string) string {
	// 适配操作系统名称
	adoptOs := osName
	if osName == "windows" {
		adoptOs = "windows"
	} else if osName == "darwin" {
		adoptOs = "mac"
	} else if osName == "linux" {
		adoptOs = "linux"
	}

	// 适配架构名称
	adoptArch := arch
	if arch == "x64" || arch == "amd64" {
		adoptArch = "x64"
	} else if arch == "x86" || arch == "386" {
		adoptArch = "x86"
	} else if arch == "arm64" {
		adoptArch = "aarch64"
	}

	// 构建API URL
	apiUrl := fmt.Sprintf(
		"https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&os=%s&image_type=jdk&vendor=eclipse",
		version, adoptArch, adoptOs,
	)

	// 获取下载链接
	resp, err := http.Get(apiUrl)
	if err != nil {
		utils.Log.Warning(fmt.Sprintf("获取下载链接失败: %v", err))
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Warning(fmt.Sprintf("读取下载链接失败: %v", err))
		return ""
	}

	var releases []struct {
		BinaryLink string `json:"binary_link"`
	}

	if err := json.Unmarshal(body, &releases); err != nil {
		utils.Log.Warning(fmt.Sprintf("解析下载链接失败: %v", err))
		return ""
	}

	if len(releases) == 0 {
		utils.Log.Warning("警告：未找到适合当前系统的Java版本")
		return ""
	}

	return releases[0].BinaryLink
}

// GetExtractDir 获取解压后的目录名
func (p *JavaSDKProvider) GetExtractDir(version, downloadedFile string) string {
	// Java SDK通常会有一个子目录，我们可以通过检查解压出来的目录结构来确定
	return "" // 在PostInstall中处理子目录
}

// GetBinDir 获取bin目录
func (p *JavaSDKProvider) GetBinDir(baseDir string) string {
	return filepath.Join(baseDir, "bin")
}

// ConfigureEnv 配置环境变量
func (p *JavaSDKProvider) ConfigureEnv(version, installDir string) ([]config.EnvVar, error) {
	// 确保目录存在
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Java安装目录不存在: %s", installDir)
	}

	// 获取bin目录
	binDir := filepath.Join(installDir, "bin")

	// 检查bin目录是否存在
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Java bin目录不存在: %s", binDir)
	}

	return []config.EnvVar{
		{
			Key:   "JAVA_HOME",
			Value: installDir,
		},
		{
			Key:   "PATH",
			Value: binDir,
		},
		{
			Key:   "EXCLUDE_KEYWORDS",
			Value: "java,jdk,openjdk",
		},
	}, nil
}

// PreInstall 安装前的准备工作
func (p *JavaSDKProvider) PreInstall(version string) error {
	// 对于Java，不需要特殊的安装前准备
	return nil
}

// PostInstall 安装后的处理工作
func (p *JavaSDKProvider) PostInstall(version, installDir string) error {
	// 查找JDK目录
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return fmt.Errorf("读取安装目录失败: %w", err)
	}

	// 查找JDK目录
	var jdkDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "jdk") {
			jdkDir = filepath.Join(installDir, entry.Name())
			break
		}
	}

	if jdkDir == "" {
		return nil // 没有找到JDK目录，可能已经是正确的结构
	}

	// 移动JDK目录中的文件到安装目录
	utils.Log.Move(fmt.Sprintf("正在移动文件从 %s 到 %s", jdkDir, installDir))

	// 读取JDK目录中的文件
	jdkEntries, err := os.ReadDir(jdkDir)
	if err != nil {
		return fmt.Errorf("读取JDK目录失败: %w", err)
	}

	// 移动文件
	for _, entry := range jdkEntries {
		src := filepath.Join(jdkDir, entry.Name())
		dst := filepath.Join(installDir, entry.Name())

		// 如果目标文件已存在，先删除
		if _, err := os.Stat(dst); err == nil {
			if err := os.RemoveAll(dst); err != nil {
				utils.Log.Warning(fmt.Sprintf("无法删除已存在的文件 %s: %v", dst, err))
				continue
			}
		}

		// 移动文件
		if err := os.Rename(src, dst); err != nil {
			utils.Log.Warning(fmt.Sprintf("移动文件失败 %s: %v，尝试复制", src, err))

			// 如果移动失败，尝试复制
			if entry.IsDir() {
				if err := utils.CopyDir(src, dst); err != nil {
					utils.Log.Warning(fmt.Sprintf("复制目录失败 %s: %v", src, err))
					continue
				}
			} else {
				if err := utils.CopyFile(src, dst); err != nil {
					utils.Log.Warning(fmt.Sprintf("复制文件失败 %s: %v", src, err))
					continue
				}
			}
		}
	}

	// 删除JDK目录
	if err := os.RemoveAll(jdkDir); err != nil {
		utils.Log.Warning(fmt.Sprintf("删除原目录失败 %s: %v", jdkDir, err))
	}

	return nil
}

// GetArchiveType 获取归档类型
func (p *JavaSDKProvider) GetArchiveType() string {
	return "zip"
}

// GetArchiveTypeForFile 根据文件名确定正确的归档类型
func (p *JavaSDKProvider) GetArchiveTypeForFile(filePath string) string {
	fileName := filepath.Base(filePath)
	if strings.HasSuffix(fileName, ".zip") {
		return "zip"
	} else if strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tgz") {
		return "tar.gz"
	} else if strings.HasSuffix(fileName, ".exe") || strings.HasSuffix(fileName, ".msi") {
		return "none"
	}
	return "zip" // 默认为zip
}

// copyFile 辅助函数，用于复制文件
func copyFile(src, dst string) error {
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
	if err != nil {
		return err
	}

	// 保持文件权限
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
