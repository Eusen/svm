package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"svm/internal/config"
	"svm/internal/utils"
)

// SDK 定义了所有语言SDK需要实现的接口
type SDK interface {
	// List 列出所有可用版本
	List() ([]string, error)

	// ListAll 列出所有可用版本（不过滤）
	ListAll() ([]string, error)

	// Install 安装指定版本
	Install(version string) error

	// Remove 删除指定版本
	Remove(version string) error

	// Use 切换到指定版本
	Use(version string) error

	// GetName 获取SDK名称
	GetName() string

	// GetCurrentVersion 获取当前使用的版本
	GetCurrentVersion() (string, error)

	// SetupEnv 设置环境变量
	SetupEnv(version string) error
}

// SDKProvider 定义了SDK的基本行为
type SDKProvider interface {
	// GetVersionList 获取可用版本列表
	GetVersionList() ([]string, error)

	// GetAllVersionList 获取所有可用版本列表（不过滤）
	GetAllVersionList() ([]string, error)

	// GetDownloadURL 获取下载URL
	GetDownloadURL(version, osName, arch string) string

	// GetExtractDir 获取解压后的目录名
	GetExtractDir(version, downloadedFile string) string

	// GetBinDir 获取bin目录
	GetBinDir(baseDir string) string

	// ConfigureEnv 配置环境变量
	ConfigureEnv(version, installDir string) ([]config.EnvVar, error)

	// PreInstall 安装前的准备工作
	PreInstall(version string) error

	// PostInstall 安装后的处理工作
	PostInstall(version, installDir string) error

	// GetArchiveType 获取归档类型
	GetArchiveType() string

	// GetArchiveTypeForFile 根据具体文件确定归档类型
	GetArchiveTypeForFile(filePath string) string
}

// VersionPrefixHandlers 定义了不同SDK的版本前缀处理逻辑
type VersionPrefixHandlers struct {
	Add    func(string) string
	Remove func(string) string
	Has    func(string) bool
}

// DefaultVersionPrefixHandlers 返回默认的版本前缀处理函数
func DefaultVersionPrefixHandlers() VersionPrefixHandlers {
	return VersionPrefixHandlers{
		Add:    func(v string) string { return v },
		Remove: func(v string) string { return v },
		Has:    func(v string) bool { return true },
	}
}

// NodeJSVersionPrefixHandlers 返回Node.js的版本前缀处理函数
func NodeJSVersionPrefixHandlers() VersionPrefixHandlers {
	return VersionPrefixHandlers{
		Add: func(v string) string {
			if !strings.HasPrefix(v, "v") {
				return "v" + v
			}
			return v
		},
		Remove: func(v string) string {
			return strings.TrimPrefix(v, "v")
		},
		Has: func(v string) bool {
			return strings.HasPrefix(v, "v")
		},
	}
}

// BaseSDK 提供基本的SDK实现
type BaseSDK struct {
	Name            string
	InstallDir      string
	Config          *config.Config
	Provider        SDKProvider
	VersionHandlers VersionPrefixHandlers
}

// NewBaseSDK 创建一个新的BaseSDK
func NewBaseSDK(name string, provider SDKProvider, handlers VersionPrefixHandlers) *BaseSDK {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("警告：加载配置失败: %v，将使用默认配置\n", err)
		cfg = &config.Config{
			InstallDir:      filepath.Join(os.Getenv("HOME"), ".svm"),
			CurrentVersions: make(map[string]string),
			SDKs:            make(map[string]config.SDKConfig),
		}
	}

	return &BaseSDK{
		Name:            name,
		InstallDir:      filepath.Join(cfg.InstallDir, name),
		Config:          cfg,
		Provider:        provider,
		VersionHandlers: handlers,
	}
}

// GetName 实现SDK接口
func (b *BaseSDK) GetName() string {
	return b.Name
}

// GetCurrentVersion 获取当前使用的版本
func (b *BaseSDK) GetCurrentVersion() (string, error) {
	version := b.Config.GetCurrentVersion(b.GetName())
	if version == "" {
		return "", fmt.Errorf("未设置当前%s版本", b.Name)
	}
	return version, nil
}

// List 统一实现的列表功能
func (b *BaseSDK) List() ([]string, error) {
	return b.Provider.GetVersionList()
}

// ListAll 统一实现的列出所有版本功能（不过滤）
func (b *BaseSDK) ListAll() ([]string, error) {
	return b.Provider.GetAllVersionList()
}

// Install 统一实现的安装功能
func (b *BaseSDK) Install(version string) error {
	// 规范化版本号
	version = b.VersionHandlers.Add(version)

	// 执行安装前的准备工作
	if err := b.Provider.PreInstall(version); err != nil {
		return err
	}

	// 获取可用的版本列表
	availableVersions, err := b.List()
	if err != nil {
		return fmt.Errorf("无法获取可用版本列表: %w", err)
	}

	fmt.Printf("获取到 %d 个%s版本\n", len(availableVersions), b.Name)

	// 查找最佳版本
	targetVersion, found := b.FindBestVersion(version, availableVersions, b.VersionHandlers)
	if !found {
		return fmt.Errorf("无法找到合适的%s版本，请检查网络连接或手动指定有效版本", b.Name)
	}

	// 准备安装目录
	versionDir, err := b.PrepareInstallDir(targetVersion)
	if err != nil {
		return err
	}

	// 检查是否有缓存文件
	cachedFilePath, hasCachedFile := b.GetCachedFile(targetVersion)

	// 构建下载URL和处理下载仅在没有缓存时进行
	archivePath := cachedFilePath
	if !hasCachedFile {
		// 获取系统和架构信息
		osName := b.GetOSName()
		arch := b.GetArchName()

		// 获取下载URL
		downloadUrl := b.Provider.GetDownloadURL(targetVersion, osName, arch)
		if downloadUrl == "" {
			return fmt.Errorf("无法为%s版本获取下载URL", targetVersion)
		}

		fmt.Printf("下载URL: %s\n", downloadUrl)

		// 下载或使用缓存
		downloadedFile, err := b.DownloadOrUseCachedFile(downloadUrl, versionDir, targetVersion, "")
		if err != nil {
			fmt.Printf("下载失败: %v\n", err)
			fmt.Println("尝试下一个版本...")
			// 尝试回退到下一个版本
			return b.FallthroughToNextVersion(targetVersion, availableVersions, b.Install, b.VersionHandlers)
		}

		archivePath = downloadedFile
	}

	// 获取归档类型并解压
	fmt.Println("正在解压文件...")
	archiveType := b.Provider.GetArchiveType()

	// 如果archiveType是"auto"，则根据实际文件确定类型
	if archiveType == "auto" {
		archiveType = b.Provider.GetArchiveTypeForFile(archivePath)
	}

	var err2 error

	if archiveType == "zip" {
		err2 = utils.ExtractZip(archivePath, versionDir)
	} else if archiveType == "tar.gz" || archiveType == "tgz" {
		err2 = utils.ExtractTarGzFile(archivePath, versionDir)
	} else if archiveType == "none" {
		// 对于不需要解压的类型（如可执行安装程序），直接复制到目标目录
		fmt.Println("无需解压，直接处理...")

		// 如果文件不在目标目录，需要复制过去
		if filepath.Dir(archivePath) != versionDir {
			destPath := filepath.Join(versionDir, filepath.Base(archivePath))
			err2 = utils.CopyFile(archivePath, destPath)
		}
	} else {
		err2 = fmt.Errorf("不支持的归档类型: %s", archiveType)
	}

	if err2 != nil {
		return fmt.Errorf("解压失败: %w", err2)
	}

	// 处理解压后的目录结构
	extractDir := b.Provider.GetExtractDir(targetVersion, archivePath)
	if extractDir != "" {
		srcDir := filepath.Join(versionDir, extractDir)
		if _, err := os.Stat(srcDir); err == nil {
			// 移动文件到根目录
			entries, err := os.ReadDir(srcDir)
			if err == nil {
				for _, entry := range entries {
					src := filepath.Join(srcDir, entry.Name())
					dst := filepath.Join(versionDir, entry.Name())

					// 检查目标文件是否已存在
					if _, err := os.Stat(dst); err == nil {
						// 如果已存在，尝试删除
						if err := os.RemoveAll(dst); err != nil {
							fmt.Printf("警告：无法删除已存在的文件 %s: %v\n", dst, err)
							continue
						}
					}

					// 尝试移动文件
					if err := os.Rename(src, dst); err != nil {
						// 如果移动失败，尝试复制
						fmt.Printf("移动文件失败，尝试复制: %s -> %s\n", src, dst)
						if utils.IsDirEntry(entry) {
							if err := utils.CopyDir(src, dst); err != nil {
								fmt.Printf("警告：复制目录失败 %s -> %s: %v\n", src, dst, err)
							}
						} else {
							if err := utils.CopyFile(src, dst); err != nil {
								fmt.Printf("警告：复制文件失败 %s -> %s: %v\n", src, dst, err)
							}
						}
					}
				}
				// 尝试删除子目录
				if err := os.RemoveAll(srcDir); err != nil {
					fmt.Printf("警告：无法删除子目录 %s: %v\n", srcDir, err)
				}
			}
		}
	}

	// 执行安装后的处理
	if err := b.Provider.PostInstall(targetVersion, versionDir); err != nil {
		return err
	}

	fmt.Printf("%s %s 安装完成\n", b.Name, targetVersion)
	return nil
}

// Remove 统一实现的移除功能
func (b *BaseSDK) Remove(version string) error {
	// 规范化版本号
	version = b.VersionHandlers.Add(version)
	return b.RemoveSDKVersion(version)
}

// Use 统一实现的切换版本功能
func (b *BaseSDK) Use(version string) error {
	// 规范化版本号
	version = b.VersionHandlers.Add(version)

	// 检查是否是已安装的版本
	versionDir := filepath.Join(b.InstallDir, version)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		// 如果指定的版本目录不存在，尝试获取匹配的版本
		fullVersion, err := b.getLatestMatchingVersion(version)
		if err != nil {
			return fmt.Errorf("获取版本信息失败: %w", err)
		}

		// 更新版本和版本目录
		version = fullVersion
		versionDir = filepath.Join(b.InstallDir, version)

		// 再次检查版本是否已安装
		if _, err := os.Stat(versionDir); os.IsNotExist(err) {
			fmt.Printf("版本 %s 未安装，正在自动安装...\n", version)
			if err := b.Install(version); err != nil {
				return err
			}
		}
	}

	// 创建或更新软链接
	currentDir := filepath.Join(b.InstallDir, "current")

	// 如果current目录已存在，先删除它
	if _, err := os.Stat(currentDir); err == nil {
		// 在Windows上，需要先检查是否是目录或符号链接
		if runtime.GOOS == "windows" {
			// 获取文件属性
			fileInfo, err := os.Lstat(currentDir)
			if err != nil {
				return fmt.Errorf("获取文件信息失败: %w", err)
			}

			// 检查是否是目录
			if fileInfo.IsDir() {
				// 如果是目录，使用os.RemoveAll删除
				if err := os.RemoveAll(currentDir); err != nil {
					return fmt.Errorf("删除现有目录失败: %w", err)
				}
			} else {
				// 如果是符号链接或文件，使用os.Remove删除
				if err := os.Remove(currentDir); err != nil {
					return fmt.Errorf("删除现有链接失败: %w", err)
				}
			}
		} else {
			// 在Unix系统上，直接删除符号链接
			if err := os.Remove(currentDir); err != nil {
				return fmt.Errorf("删除现有链接失败: %w", err)
			}
		}
	}

	// 创建从current到版本目录的符号链接
	if runtime.GOOS == "windows" {
		// Windows需要管理员权限创建符号链接，使用junction作为替代
		// 使用mklink命令创建目录连接
		cmd := exec.Command("cmd", "/c", "mklink", "/J", currentDir, versionDir)
		if err := cmd.Run(); err != nil {
			// 如果mklink失败，尝试使用复制作为后备方案
			fmt.Printf("警告：创建目录连接失败，将使用复制作为替代方案: %v\n", err)
			if err := utils.CopyDir(versionDir, currentDir); err != nil {
				return fmt.Errorf("复制目录失败: %w", err)
			}
		}
	} else {
		// Unix系统直接创建符号链接
		if err := os.Symlink(versionDir, currentDir); err != nil {
			return fmt.Errorf("创建符号链接失败: %w", err)
		}
	}

	// 检查是否已经设置过环境变量
	sdkConfig, ok := b.Config.SDKs[b.GetName()]
	if !ok || len(sdkConfig.EnvVars) == 0 {
		// 如果没有设置过环境变量，则设置
		if err := b.SetupEnv(version); err != nil {
			return err
		}
	} else {
		// 如果已经设置过环境变量，只更新当前版本
		if err := b.Config.SetCurrentVersion(b.GetName(), version); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		// 始终更新环境变量，确保current目录被正确添加到PATH
		if err := b.SetupEnv(version); err != nil {
			return err
		}

		fmt.Printf("已切换到 %s %s\n", b.Name, version)
	}

	return nil
}

// SetupEnv 设置环境变量
func (b *BaseSDK) SetupEnv(version string) error {
	// 使用固定的current目录而不是版本目录
	currentDir := filepath.Join(b.InstallDir, "current")

	// 确保current目录存在
	if _, err := os.Stat(currentDir); os.IsNotExist(err) {
		return fmt.Errorf("current目录不存在，请先使用use命令切换版本")
	}

	// 获取环境变量配置
	envVars, err := b.Provider.ConfigureEnv(version, currentDir)
	if err != nil {
		return err
	}

	// 保存环境变量配置
	if err := b.Config.SetSDKEnvVars(b.GetName(), envVars); err != nil {
		return fmt.Errorf("保存环境变量配置失败: %w", err)
	}

	// 获取主环境变量和PATH
	var homeVar, homePath, binPath string
	var excludeKeywords []string
	extraVars := make(map[string]string)

	for _, env := range envVars {
		if strings.HasSuffix(env.Key, "_HOME") {
			homeVar = env.Key
			homePath = env.Value
		} else if env.Key == "PATH" {
			binPath = env.Value
		} else if env.Key == "EXCLUDE_KEYWORDS" {
			excludeKeywords = strings.Split(env.Value, ",")
		} else if env.Key != "" && env.Value != "" {
			// 处理其他环境变量
			extraVars[env.Key] = env.Value
		}
	}

	// 如果没有指定bin路径，使用provider提供的
	if binPath == "" {
		binPath = b.Provider.GetBinDir(currentDir)
	}

	// 使用环境变量管理器设置环境变量
	envManager := &utils.EnvManager{
		Name:            b.Name,
		HomeVar:         homeVar,
		HomePath:        homePath,
		BinPath:         binPath,
		ExcludeKeywords: excludeKeywords,
		ExtraVars:       extraVars,
	}

	if err := envManager.SetEnv(version); err != nil {
		return err
	}

	// 保存当前版本到配置文件
	if err := b.Config.SetCurrentVersion(b.GetName(), version); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	return nil
}

// FindBestVersion 查找最佳匹配的版本
// 如果请求的版本不存在，则尝试找到最接近的较低版本
func (b *BaseSDK) FindBestVersion(requestedVersion string, availableVersions []string, handlers VersionPrefixHandlers) (string, bool) {
	// 标准化版本号格式
	originalVersion := requestedVersion
	requestedVersion = handlers.Add(requestedVersion)

	// 确保版本列表已经按降序排序（从新到旧）
	utils.SortVersionsDesc(availableVersions)

	// 调试输出
	if len(availableVersions) > 0 {
		count := min(5, len(availableVersions))
		fmt.Printf("最新的几个%s版本: %v\n", b.Name, availableVersions[:count])
	}

	// 使用utils包中的函数查找最佳匹配版本
	// 构造stripPrefix参数 - 对于Node.js是"v"，对于其他SDK是""
	stripPrefix := ""
	if handlers.Has(requestedVersion) {
		stripPrefix = requestedVersion[:1] // 获取前缀的第一个字符
	}

	targetVersion, found := utils.FindBestMatchingVersion(requestedVersion, availableVersions, stripPrefix)

	if found && targetVersion != requestedVersion {
		fmt.Printf("请求的版本 %s 不可用，将使用 %s\n", originalVersion, targetVersion)
	} else if found {
		fmt.Printf("找到匹配的版本: %s\n", targetVersion)
	}

	return targetVersion, found
}

// ValidateDownloadURL 验证下载URL是否有效
func (b *BaseSDK) ValidateDownloadURL(url string) (bool, error) {
	fmt.Printf("验证下载URL: %s\n", url)
	exists, err := utils.CheckURLExists(url)
	if err != nil {
		fmt.Printf("验证URL失败: %v\n", err)
	} else {
		if exists {
			fmt.Printf("URL有效\n")
		} else {
			fmt.Printf("URL无效\n")
		}
	}
	return exists, err
}

// PrepareInstallDir 准备安装目录，优先检查是否已有安装目录
func (b *BaseSDK) PrepareInstallDir(version string) (string, error) {
	// 检查配置中是否有版本信息
	versionInfo, exists := b.Config.GetVersionInfo(b.GetName(), version)

	// 已有安装目录，直接返回
	if exists && versionInfo.InstallDir != "" {
		// 验证目录是否存在
		if _, err := os.Stat(versionInfo.InstallDir); err == nil {
			fmt.Printf("发现已有安装目录: %s\n", versionInfo.InstallDir)
			return versionInfo.InstallDir, nil
		}
	}

	// 创建安装目录
	versionDir := filepath.Join(b.InstallDir, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return "", fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 清理可能存在的旧文件
	entries, err := os.ReadDir(versionDir)
	if err == nil && len(entries) > 0 {
		fmt.Println("清理安装目录中的旧文件...")
		for _, entry := range entries {
			path := filepath.Join(versionDir, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("警告：无法删除文件 %s: %v\n", path, err)
			}
		}
	}

	// 保存安装目录信息到配置
	if !exists {
		// 如果不存在版本信息，创建新的，保留可能存在的缓存文件路径
		versionInfo = config.SDKVersionInfo{InstallDir: versionDir}
	} else {
		// 更新安装目录，保留缓存文件路径
		versionInfo.InstallDir = versionDir
	}

	// 保存到配置
	if err := b.Config.SetVersionInfo(b.GetName(), version, versionInfo); err != nil {
		fmt.Printf("警告：保存版本信息失败: %v\n", err)
	}

	return versionDir, nil
}

// GetArchName 获取当前架构名称
func (b *BaseSDK) GetArchName() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		return "x64"
	} else if arch == "386" {
		return "x86"
	} else if arch == "arm64" {
		return "arm64"
	}
	return arch
}

// GetOSName 获取当前操作系统名称
func (b *BaseSDK) GetOSName() string {
	osName := runtime.GOOS
	return osName
}

// CleanupTempFile 清理临时文件
func (b *BaseSDK) CleanupTempFile(filePath string) {
	if err := os.Remove(filePath); err != nil {
		fmt.Printf("警告：清理临时文件失败: %v\n", err)
	} else {
		fmt.Printf("清理临时文件: %s\n", filePath)
	}
}

// FallthroughToNextVersion 尝试下一个可用版本
func (b *BaseSDK) FallthroughToNextVersion(currentVersion string, availableVersions []string, installer func(string) error, handlers VersionPrefixHandlers) error {
	// 清理当前版本的安装目录
	currentVersionDir := filepath.Join(b.InstallDir, currentVersion)
	if _, err := os.Stat(currentVersionDir); err == nil {
		fmt.Printf("清理失败的安装目录: %s\n", currentVersionDir)
		if err := os.RemoveAll(currentVersionDir); err != nil {
			fmt.Printf("警告：无法删除目录 %s: %v\n", currentVersionDir, err)
		}
	}

	// 从配置中移除当前版本信息
	if err := b.Config.RemoveVersionInfo(b.GetName(), currentVersion); err != nil {
		fmt.Printf("警告：无法从配置中移除版本信息: %v\n", err)
	}

	nextVersion, found := utils.GetNextVersionFromList(currentVersion, availableVersions)
	if found {
		fmt.Printf("版本 %s 无法下载，尝试下一个可用版本: %s\n", currentVersion, nextVersion)

		// 递归尝试下一个版本
		return installer(handlers.Remove(nextVersion))
	}

	return fmt.Errorf("无法找到可下载的%s版本，请检查网络连接或手动指定有效版本", b.Name)
}

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// SaveCacheFile 保存缓存文件信息
func (b *BaseSDK) SaveCacheFile(version string, cacheFilePath string) error {
	// 获取当前版本信息
	versionInfo, exists := b.Config.GetVersionInfo(b.GetName(), version)

	if !exists {
		// 如果不存在版本信息，创建新的
		versionInfo = config.SDKVersionInfo{
			CacheFilePath: cacheFilePath,
		}
	} else {
		// 更新缓存文件路径
		versionInfo.CacheFilePath = cacheFilePath
	}

	// 保存到配置
	return b.Config.SetVersionInfo(b.GetName(), version, versionInfo)
}

// GetCachedFile 获取缓存文件路径
func (b *BaseSDK) GetCachedFile(version string) (string, bool) {
	versionInfo, exists := b.Config.GetVersionInfo(b.GetName(), version)
	if !exists || versionInfo.CacheFilePath == "" {
		return "", false
	}

	// 检查文件是否存在
	if _, err := os.Stat(versionInfo.CacheFilePath); os.IsNotExist(err) {
		fmt.Printf("已记录缓存文件路径，但文件不存在: %s\n", versionInfo.CacheFilePath)
		return "", false
	}

	fmt.Printf("使用缓存文件: %s (跳过下载)\n", versionInfo.CacheFilePath)
	return versionInfo.CacheFilePath, true
}

// DownloadOrUseCachedFile 下载文件或使用缓存文件
func (b *BaseSDK) DownloadOrUseCachedFile(url string, targetDir string, version string, tip string) (string, error) {
	// 检查是否有缓存文件
	cachedFilePath, hasCachedFile := b.GetCachedFile(version)
	if hasCachedFile {
		return cachedFilePath, nil
	}

	// 没有缓存文件，下载新文件
	fileName := filepath.Base(url)

	// 创建缓存目录
	cacheDir := filepath.Join(b.Config.GetCacheDir(), b.GetName())
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 缓存文件路径
	filePath := filepath.Join(cacheDir, fileName)

	fmt.Printf("下载文件: %s\n", url)
	fmt.Printf("缓存路径: %s\n", filePath)

	if tip != "" {
		fmt.Printf("%s\n", tip)
	}

	if err := utils.DownloadFile(url, filePath); err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}

	// 保存缓存文件信息
	if err := b.SaveCacheFile(version, filePath); err != nil {
		fmt.Printf("警告：保存缓存文件信息失败: %v\n", err)
	}

	return filePath, nil
}

// RemoveSDKVersion 删除SDK版本
func (b *BaseSDK) RemoveSDKVersion(version string) error {
	// 获取版本信息
	versionInfo, exists := b.Config.GetVersionInfo(b.GetName(), version)
	if !exists || versionInfo.InstallDir == "" {
		return fmt.Errorf("版本 %s 未安装", version)
	}

	// 检查是否为当前版本
	currentVersion := b.Config.GetCurrentVersion(b.GetName())
	if currentVersion == version {
		// 如果是当前版本，清除环境变量设置
		if err := b.Config.SetSDKEnvVars(b.GetName(), []config.EnvVar{}); err != nil {
			fmt.Printf("警告：清除环境变量失败: %v\n", err)
		}

		// 清除当前版本
		if err := b.Config.SetCurrentVersion(b.GetName(), ""); err != nil {
			fmt.Printf("警告：清除当前版本失败: %v\n", err)
		}
	}

	// 删除安装目录
	if err := os.RemoveAll(versionInfo.InstallDir); err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}

	// 将安装目录置空，但保留缓存文件信息
	versionInfo.InstallDir = ""
	if err := b.Config.SetVersionInfo(b.GetName(), version, versionInfo); err != nil {
		fmt.Printf("警告：更新版本信息失败: %v\n", err)
	}

	fmt.Printf("%s %s 已删除\n", b.GetName(), version)
	return nil
}

// getLatestMatchingVersion 获取最新匹配的版本
func (b *BaseSDK) getLatestMatchingVersion(versionPrefix string) (string, error) {
	versions, err := b.List()
	if err != nil {
		return "", err
	}

	// 处理空版本，返回最新版本
	if versionPrefix == "" || versionPrefix == b.VersionHandlers.Add("") {
		if len(versions) > 0 {
			return versions[0], nil
		}
		return "", fmt.Errorf("没有可用的%s版本", b.Name)
	}

	// 寻找匹配的版本
	for _, v := range versions {
		if strings.HasPrefix(v, versionPrefix) {
			return v, nil
		}
	}

	// 尝试寻找近似匹配
	targetVersion, found := b.FindBestVersion(versionPrefix, versions, b.VersionHandlers)
	if found {
		return targetVersion, nil
	}

	return "", fmt.Errorf("没有找到匹配的%s版本: %s", b.Name, versionPrefix)
}
