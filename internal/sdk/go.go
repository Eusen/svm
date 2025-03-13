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

// GoVersion 表示Go版本信息
type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

// GoSDKProvider 实现了SDKProvider接口
type GoSDKProvider struct {
	config *config.Config
}

// goSDK 是Go SDK的具体实现
type goSDK struct {
	BaseSDK
}

// NewGoSDK 创建一个新的Go SDK
func NewGoSDK() SDK {
	provider := &GoSDKProvider{
		config: nil, // 这里为空，会由BaseSDK初始化时设置
	}

	return &goSDK{
		BaseSDK: *NewBaseSDK("go", provider, DefaultVersionPrefixHandlers()),
	}
}

// GetVersionList 实现SDKProvider接口，获取所有可用的Go版本
func (p *GoSDKProvider) GetVersionList() ([]string, error) {
	// 从Go官网API获取版本列表
	resp, err := http.Get("https://go.dev/dl/?mode=json&include=all")
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var versions []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}

	// 提取版本号，按次版本分组
	minorVersions := make(map[string]string)
	for _, v := range versions {
		if v.Stable {
			version := v.Version
			if strings.HasPrefix(version, "go") {
				version = version[2:] // 移除"go"前缀
			}

			// 提取次版本号（例如：1.20.3 -> 1.20）
			parts := strings.Split(version, ".")
			if len(parts) >= 2 {
				minorVersion := parts[0] + "." + parts[1]

				// 如果该次版本尚未记录或当前版本更新，则更新
				if currentVersion, exists := minorVersions[minorVersion]; !exists {
					minorVersions[minorVersion] = version
				} else {
					// 比较版本号
					if version > currentVersion {
						minorVersions[minorVersion] = version
					}
				}
			} else {
				// 对于格式不符合预期的版本，直接添加
				minorVersions[version] = version
			}
		}
	}

	// 将次版本映射转换为版本列表
	var versionList []string
	for _, version := range minorVersions {
		versionList = append(versionList, version)
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versionList)

	return versionList, nil
}

// GetAllVersionList 实现SDKProvider接口，获取所有可用的Go版本（不过滤）
func (p *GoSDKProvider) GetAllVersionList() ([]string, error) {
	// 从Go官网API获取版本列表
	resp, err := http.Get("https://go.dev/dl/?mode=json&include=all")
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var versions []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}

	// 提取所有稳定版本号
	var versionList []string
	for _, v := range versions {
		if v.Stable {
			version := v.Version
			if strings.HasPrefix(version, "go") {
				version = version[2:] // 移除"go"前缀
			}
			versionList = append(versionList, version)
		}
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versionList)

	return versionList, nil
}

// GetDownloadURL 构建Go下载URL
func (p *GoSDKProvider) GetDownloadURL(version, osName, arch string) string {
	// 适配操作系统名称
	goOs := osName
	if osName == "darwin" {
		goOs = "darwin"
	} else if osName == "windows" {
		goOs = "windows"
	} else if osName == "linux" {
		goOs = "linux"
	}

	// 适配架构名称
	goArch := arch
	if arch == "x64" || arch == "amd64" {
		goArch = "amd64"
	} else if arch == "x86" || arch == "386" {
		goArch = "386"
	} else if arch == "arm64" {
		goArch = "arm64"
	}

	// 确定文件扩展名
	ext := "tar.gz"
	if goOs == "windows" {
		ext = "zip"
	}

	// 构建下载URL
	return fmt.Sprintf("https://dl.google.com/go/go%s.%s-%s.%s", version, goOs, goArch, ext)
}

// GetExtractDir 获取解压后的目录名
func (p *GoSDKProvider) GetExtractDir(version, downloadedFile string) string {
	// Go的解压目录是 "go"
	return "go"
}

// GetBinDir 获取bin目录
func (p *GoSDKProvider) GetBinDir(baseDir string) string {
	return filepath.Join(baseDir, "bin")
}

// ConfigureEnv 配置环境变量
func (p *GoSDKProvider) ConfigureEnv(version, installDir string) ([]config.EnvVar, error) {
	// 确保目录存在
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Go安装目录不存在: %s", installDir)
	}

	// 获取bin目录
	binDir := filepath.Join(installDir, "bin")

	// 检查bin目录是否存在
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Go bin目录不存在: %s", binDir)
	}

	return []config.EnvVar{
		{
			Key:   "GOROOT",
			Value: installDir,
		},
		{
			Key:   "PATH",
			Value: binDir,
		},
		{
			Key:   "EXCLUDE_KEYWORDS",
			Value: "golang,go",
		},
	}, nil
}

// PreInstall 安装前的准备工作
func (p *GoSDKProvider) PreInstall(version string) error {
	// 对于Go，不需要特殊的安装前准备
	return nil
}

// PostInstall 安装后的处理工作
func (p *GoSDKProvider) PostInstall(version, installDir string) error {
	// 对于Go，我们需要处理可能存在的go目录
	goDir := filepath.Join(installDir, "go")
	if _, err := os.Stat(goDir); err == nil {
		// 移动文件
		entries, err := os.ReadDir(goDir)
		if err != nil {
			return fmt.Errorf("读取go目录失败: %w", err)
		}

		for _, entry := range entries {
			src := filepath.Join(goDir, entry.Name())
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
				utils.Log.Warning(fmt.Sprintf("移动文件失败 %s: %v", src, err))
			}
		}

		// 删除go目录
		if err := os.RemoveAll(goDir); err != nil {
			utils.Log.Warning(fmt.Sprintf("删除go目录失败: %v", err))
		}
	}

	return nil
}

// GetArchiveType 获取归档类型
func (p *GoSDKProvider) GetArchiveType() string {
	return "zip"
}

// GetArchiveTypeForFile 根据文件名确定正确的归档类型
func (p *GoSDKProvider) GetArchiveTypeForFile(filePath string) string {
	fileName := filepath.Base(filePath)
	if strings.HasSuffix(fileName, ".zip") {
		return "zip"
	} else if strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tgz") {
		return "tar.gz"
	}
	return "zip" // 默认为zip
}
