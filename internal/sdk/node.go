package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"svm/internal/config"
	"svm/internal/utils"
)

// NodeVersion 表示Node.js版本信息
type NodeVersion struct {
	Version string   `json:"version"`
	Date    string   `json:"date"`
	Files   []string `json:"files"`
}

// NodeSDKProvider 实现了SDKProvider接口
type NodeSDKProvider struct {
	config *config.Config
}

// NewNodeSDK 创建一个新的Node.js SDK
func NewNodeSDK() SDK {
	provider := &NodeSDKProvider{
		config: nil, // 这里为空，会由BaseSDK初始化时设置
	}

	return &nodeSDK{
		BaseSDK: *NewBaseSDK("node", provider, NodeJSVersionPrefixHandlers()),
	}
}

// nodeSDK 是Node.js SDK的具体实现
type nodeSDK struct {
	BaseSDK
}

// GetCurrentVersion 获取当前使用的Node.js版本
func (s *nodeSDK) GetCurrentVersion() (string, error) {
	// 使用BaseSDK的GetCurrentVersion方法
	version, err := s.BaseSDK.GetCurrentVersion()
	if err != nil {
		return "", nil // 未设置版本
	}
	return version, nil
}

// GetVersionList 实现SDKProvider接口，获取所有可用的Node.js版本
func (p *NodeSDKProvider) GetVersionList() ([]string, error) {
	// 从Node.js官网获取版本列表
	resp, err := http.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var versions []NodeVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}

	// 提取版本号，并按主版本分组
	majorVersions := make(map[string]string)
	for _, v := range versions {
		// 提取版本号
		versionStr := v.Version

		// 提取主版本号（例如：v12.9.1 -> v12）
		parts := strings.Split(strings.TrimPrefix(versionStr, "v"), ".")
		if len(parts) >= 1 {
			majorVersion := "v" + parts[0]

			// 如果该主版本尚未记录或当前版本更新，则更新
			if currentVersion, exists := majorVersions[majorVersion]; !exists {
				majorVersions[majorVersion] = versionStr
			} else {
				// 使用utils.CompareVersionsStr函数进行版本比较
				v1 := strings.TrimPrefix(versionStr, "v")
				v2 := strings.TrimPrefix(currentVersion, "v")

				if utils.CompareVersionsStr(v1, v2) > 0 {
					majorVersions[majorVersion] = versionStr
				}
			}
		}
	}

	// 将主版本映射转换为版本列表
	var versionList []string
	for _, version := range majorVersions {
		versionList = append(versionList, version)
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versionList)

	return versionList, nil
}

// GetAllVersionList 实现SDKProvider接口，获取所有可用的Node.js版本（不过滤）
func (p *NodeSDKProvider) GetAllVersionList() ([]string, error) {
	// 从Node.js官网获取版本列表
	resp, err := http.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var versions []NodeVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}

	// 提取所有版本号
	var versionList []string
	for _, v := range versions {
		versionList = append(versionList, v.Version)
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versionList)

	return versionList, nil
}

// GetDownloadURL 构建Node.js下载URL
func (p *NodeSDKProvider) GetDownloadURL(version, osName, arch string) string {
	// 根据操作系统调整名称
	if osName == "windows" {
		osName = "win"
	} else if osName == "darwin" {
		osName = "darwin"
	} else if osName == "linux" {
		osName = "linux"
	}

	// 构建ZIP文件名和下载URL
	zipFileName := fmt.Sprintf("node-%s-%s-%s.zip", version, osName, arch)
	return fmt.Sprintf("https://nodejs.org/dist/%s/%s", version, zipFileName)
}

// GetExtractDir 获取解压后的目录名
func (p *NodeSDKProvider) GetExtractDir(version, downloadedFile string) string {
	// 获取操作系统和架构
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 根据操作系统调整名称
	if osName == "windows" {
		osName = "win"
	} else if osName == "darwin" {
		osName = "darwin"
	} else if osName == "linux" {
		osName = "linux"
	}

	// 对于arm64架构
	if arch == "arm64" {
		arch = "arm64"
	} else if arch == "amd64" {
		arch = "x64"
	} else if arch == "386" {
		arch = "x86"
	}

	// 返回解压后的目录名
	return fmt.Sprintf("node-%s-%s-%s", version, osName, arch)
}

// GetBinDir 获取bin目录
func (p *NodeSDKProvider) GetBinDir(baseDir string) string {
	return baseDir
}

// ConfigureEnv 配置环境变量
func (p *NodeSDKProvider) ConfigureEnv(version, installDir string) ([]config.EnvVar, error) {
	// Node.js只需要设置PATH
	return []config.EnvVar{
		{
			Key:   "NODE_HOME",
			Value: installDir,
		},
		{
			Key:   "EXCLUDE_KEYWORDS",
			Value: "node",
		},
	}, nil
}

// PreInstall 安装前的准备工作
func (p *NodeSDKProvider) PreInstall(version string) error {
	// 对于Node.js，不需要特殊的安装前准备
	return nil
}

// PostInstall 安装后的处理工作
func (p *NodeSDKProvider) PostInstall(version, installDir string) error {
	// 对于Node.js，不需要特殊的安装后处理
	return nil
}

// GetArchiveType 获取归档类型
func (p *NodeSDKProvider) GetArchiveType() string {
	return "zip"
}

// GetArchiveTypeForFile 根据文件名确定正确的归档类型
func (p *NodeSDKProvider) GetArchiveTypeForFile(filePath string) string {
	fileName := filepath.Base(filePath)
	if strings.HasSuffix(fileName, ".zip") {
		return "zip"
	} else if strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tgz") {
		return "tar.gz"
	}
	return "zip" // 默认为zip
}
