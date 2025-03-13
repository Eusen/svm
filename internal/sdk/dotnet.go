package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"svm/internal/config"
	"svm/internal/utils"
)

// DotNetReleasesIndex 表示.NET版本索引信息
type DotNetReleasesIndex struct {
	ReleasesIndex []DotNetReleaseInfo `json:"releases-index"`
}

// DotNetReleaseInfo 表示.NET版本信息
type DotNetReleaseInfo struct {
	ChannelVersion    string `json:"channel-version"`
	LatestRelease     string `json:"latest-release"`
	LatestReleaseDate string `json:"latest-release-date"`
	Security          bool   `json:"security"`
	LatestRuntime     string `json:"latest-runtime"`
	LatestSDK         string `json:"latest-sdk"`
	Product           string `json:"product"`
	SupportPhase      string `json:"support-phase"`
	ReleaseType       string `json:"release-type"`
	ReleasesJSON      string `json:"releases.json"`
}

// DotNetReleasesJSON 表示.NET版本详细信息
type DotNetReleasesJSON struct {
	Releases []DotNetReleaseDetail `json:"releases"`
}

// DotNetReleaseDetail 表示.NET版本详细信息
type DotNetReleaseDetail struct {
	ReleaseVersion string                `json:"release-version"`
	ChannelVersion string                `json:"channel-version"`
	ReleaseDate    string                `json:"release-date"`
	Runtime        DotNetComponentInfo   `json:"runtime"`
	SDK            DotNetComponentInfo   `json:"sdk"`
	AspNetCore     DotNetComponentInfo   `json:"aspnetcore-runtime"`
	WindowsDesktop DotNetComponentInfo   `json:"windowsdesktop"`
	Files          []DotNetComponentFile `json:"files"`
}

// DotNetComponentInfo 表示.NET组件信息
type DotNetComponentInfo struct {
	Version string                `json:"version"`
	Files   []DotNetComponentFile `json:"files,omitempty"`
}

// DotNetComponentFile 表示.NET组件文件信息
type DotNetComponentFile struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	Hash          string `json:"hash"`
	HashAlgorithm string `json:"hash-algorithm"`
	RID           string `json:"rid,omitempty"`
}

// DotNetSDKProvider 实现了SDKProvider接口
type DotNetSDKProvider struct {
	config        *config.Config
	componentType string // 组件类型：sdk, runtime, asp-core, desktop
}

// NewDotNetSDK 创建一个新的.NET SDK
func NewDotNetSDK() SDK {
	provider := &DotNetSDKProvider{
		config:        nil,   // 这里为空，会由BaseSDK初始化时设置
		componentType: "sdk", // 默认为SDK
	}

	return &dotNetSDK{
		BaseSDK: *NewBaseSDK("dotnet", provider, DefaultVersionPrefixHandlers()),
	}
}

// dotNetSDK 是.NET SDK的具体实现
type dotNetSDK struct {
	BaseSDK
}

// SetComponentType 设置组件类型
func (s *dotNetSDK) SetComponentType(componentType string) {
	if provider, ok := s.Provider.(*DotNetSDKProvider); ok {
		provider.componentType = componentType
	}
}

// GetCurrentVersion 获取当前使用的.NET版本
func (s *dotNetSDK) GetCurrentVersion() (string, error) {
	// 获取Provider
	provider, ok := s.Provider.(*DotNetSDKProvider)
	if !ok {
		return "", fmt.Errorf("无效的Provider类型")
	}

	// 从配置中获取当前版本
	sdkConfig, exists := s.Config.SDKs[s.GetName()]
	if !exists {
		return "", fmt.Errorf("未设置当前%s版本", s.Name)
	}

	// 获取组件当前版本
	version, exists := sdkConfig.Components[provider.componentType]
	if !exists {
		return "", fmt.Errorf("未设置当前%s %s版本", s.Name, provider.componentType)
	}

	return version, nil
}

// SetupEnv 为 dotNetSDK 重写 SetupEnv 方法，确保每个组件有自己的 current 目录
func (s *dotNetSDK) SetupEnv(version string) error {
	// 获取Provider
	provider, ok := s.Provider.(*DotNetSDKProvider)
	if !ok {
		return fmt.Errorf("无效的Provider类型")
	}

	// 构建组件目录和版本目录
	componentDir := filepath.Join(s.InstallDir, provider.componentType)
	versionDir := filepath.Join(componentDir, version)
	currentDir := filepath.Join(componentDir, "current")

	// 检查版本目录是否存在
	exists, err := utils.CheckDirExists(versionDir)
	if err != nil || !exists {
		return fmt.Errorf("版本目录不存在: %s", versionDir)
	}

	// 删除旧的current目录或符号链接
	if fileInfo, err := os.Lstat(currentDir); err == nil {
		// 检查是否是符号链接
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			utils.Log.Delete(fmt.Sprintf("正在删除旧的符号链接: %s", currentDir))
			if err := os.Remove(currentDir); err != nil {
				utils.Log.Error(fmt.Sprintf("删除旧的符号链接失败: %v", err))
				return fmt.Errorf("删除旧的符号链接失败: %w", err)
			}
		} else {
			// 是目录，删除它
			utils.Log.Delete(fmt.Sprintf("正在删除旧的 current 目录: %s", currentDir))
			if err := os.RemoveAll(currentDir); err != nil {
				utils.Log.Error(fmt.Sprintf("删除旧的 current 目录失败: %v", err))
				return fmt.Errorf("删除旧的current目录失败: %w", err)
			}
		}
	}

	// 创建从current到版本目录的符号链接
	if runtime.GOOS == "windows" {
		// Windows需要管理员权限创建符号链接，使用junction作为替代
		// 使用mklink命令创建目录连接
		utils.Log.Link(fmt.Sprintf("正在创建目录连接: %s -> %s", currentDir, versionDir))
		cmd := exec.Command("cmd", "/c", "mklink", "/J", currentDir, versionDir)
		if err := cmd.Run(); err != nil {
			// 如果mklink失败，尝试使用复制作为后备方案
			utils.Log.Warning(fmt.Sprintf("创建目录连接失败，将使用复制作为替代方案: %v", err))
			if err := utils.CopyDir(versionDir, currentDir); err != nil {
				return fmt.Errorf("复制目录失败: %w", err)
			}
		}
	} else {
		// Unix系统直接创建符号链接
		utils.Log.Link(fmt.Sprintf("正在创建符号链接: %s -> %s", currentDir, versionDir))
		if err := os.Symlink(versionDir, currentDir); err != nil {
			return fmt.Errorf("创建符号链接失败: %w", err)
		}
	}

	// 创建一个文件来记录当前版本
	versionFile := filepath.Join(currentDir, ".version")
	if err := os.WriteFile(versionFile, []byte(version), 0644); err != nil {
		utils.Log.Warning(fmt.Sprintf("写入版本文件失败: %v", err))
	}

	// 获取环境变量配置
	envVars, err := provider.ConfigureEnv(version, componentDir)
	if err != nil {
		return err
	}

	// 保存环境变量配置
	if err := s.Config.SetSDKEnvVars(s.GetName(), envVars); err != nil {
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
		binPath = provider.GetBinDir(currentDir)
	}

	// 使用环境变量管理器设置环境变量
	envManager := &utils.EnvManager{
		Name:            s.Name,
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
	sdkConfig, exists := s.Config.SDKs[s.GetName()]
	if !exists {
		sdkConfig = config.SDKConfig{
			Components:   make(map[string]string),
			VersionCache: make(map[string]config.SDKVersionInfo),
		}
	}

	// 更新组件当前版本
	if sdkConfig.Components == nil {
		sdkConfig.Components = make(map[string]string)
	}
	sdkConfig.Components[provider.componentType] = version

	// 保存配置
	s.Config.SDKs[s.GetName()] = sdkConfig
	if err := s.Config.Save(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	utils.Log.Config(fmt.Sprintf("已设置 %s %s %s 环境变量", s.Name, provider.componentType, version))
	return nil
}

// Use 切换到指定版本
func (s *dotNetSDK) Use(version string) error {
	// 获取Provider
	provider, ok := s.Provider.(*DotNetSDKProvider)
	if !ok {
		return fmt.Errorf("无效的Provider类型")
	}

	// 检查版本是否已安装
	versionDir := filepath.Join(s.InstallDir, provider.componentType, version)
	utils.Log.Check(fmt.Sprintf("检查版本目录: %s", versionDir))

	exists, err := utils.CheckDirExists(versionDir)
	if err != nil || !exists {
		utils.Log.Warning(fmt.Sprintf("版本目录不存在: %s", versionDir))
		utils.Log.Install(fmt.Sprintf("%s %s版本 %s 未安装，正在自动安装...", s.Name, provider.componentType, version))

		// 自动安装该版本
		if err := s.Install(version); err != nil {
			return fmt.Errorf("安装失败: %w", err)
		}

		// 安装成功后，重新检查版本目录
		exists, err = utils.CheckDirExists(versionDir)
		if err != nil || !exists {
			return fmt.Errorf("安装后版本目录仍不存在: %s", versionDir)
		}
	}

	utils.Log.Success(fmt.Sprintf("找到版本目录: %s", versionDir))

	// 设置环境变量
	if err := s.SetupEnv(version); err != nil {
		return fmt.Errorf("设置环境变量失败: %w", err)
	}

	// 更新配置
	sdkConfig, exists := s.Config.SDKs[s.GetName()]
	if !exists {
		sdkConfig = config.SDKConfig{
			Components:   make(map[string]string),
			VersionCache: make(map[string]config.SDKVersionInfo),
		}
	}

	// 更新组件当前版本
	if sdkConfig.Components == nil {
		sdkConfig.Components = make(map[string]string)
	}
	sdkConfig.Components[provider.componentType] = version

	// 保存配置
	s.Config.SDKs[s.GetName()] = sdkConfig
	if err := s.Config.Save(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	utils.Log.Switch(fmt.Sprintf("已切换到 %s %s %s", s.Name, provider.componentType, version))
	return nil
}

// 获取微软官方版本列表
func (p *DotNetSDKProvider) getOfficialVersions() ([]DotNetReleaseInfo, error) {
	// 获取版本索引
	data, err := utils.FetchJSON("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/releases-index.json")
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}

	var releasesIndex DotNetReleasesIndex
	if err := json.Unmarshal(data, &releasesIndex); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}

	// 筛选出support-phase为"preview"或"active"的版本
	var filteredReleases []DotNetReleaseInfo
	for _, release := range releasesIndex.ReleasesIndex {
		if release.SupportPhase == "preview" || release.SupportPhase == "active" {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

// 获取微软所有官方版本列表
func (p *DotNetSDKProvider) getAllOfficialVersions() ([]DotNetReleaseDetail, error) {
	// 获取官方版本列表
	releases, err := p.getOfficialVersions()
	if err != nil {
		return nil, err
	}

	// 整理出releases.json URL
	var releasesJSONURLs []string
	for _, release := range releases {
		releasesJSONURLs = append(releasesJSONURLs, release.ReleasesJSON)
	}

	// 获取所有releases.json数据
	var allReleases []DotNetReleaseDetail
	for _, url := range releasesJSONURLs {
		data, err := utils.FetchJSON(url)
		if err != nil {
			utils.Log.Warning(fmt.Sprintf("获取 %s 失败: %v", url, err))
			continue
		}

		var releasesJSON DotNetReleasesJSON
		if err := json.Unmarshal(data, &releasesJSON); err != nil {
			utils.Log.Warning(fmt.Sprintf("解析 %s 失败: %v", url, err))
			continue
		}

		// 添加到总列表
		allReleases = append(allReleases, releasesJSON.Releases...)
	}

	return allReleases, nil
}

// GetVersionList 实现SDKProvider接口，获取所有可用的.NET版本
func (p *DotNetSDKProvider) GetVersionList() ([]string, error) {
	// 获取官方版本列表
	releases, err := p.getOfficialVersions()
	if err != nil {
		return nil, err
	}

	// 根据组件类型返回不同的版本列表
	var versions []string
	for _, release := range releases {
		versions = append(versions, release.LatestRelease)
	}

	return versions, nil
}

// GetAllVersionList 实现SDKProvider接口，获取所有可用的.NET版本（不过滤）
func (p *DotNetSDKProvider) GetAllVersionList() ([]string, error) {
	// 获取所有官方版本列表
	releases, err := p.getAllOfficialVersions()
	if err != nil {
		return nil, err
	}

	// 根据组件类型返回不同的版本列表
	var versions []string
	for _, release := range releases {
		versions = append(versions, release.ReleaseVersion)
	}

	return versions, nil
}

// GetDownloadURL 实现SDKProvider接口，获取下载URL
func (p *DotNetSDKProvider) GetDownloadURL(version, osName, arch string) string {
	// 获取所有官方版本列表
	releases, err := p.getAllOfficialVersions()
	if err != nil {
		utils.Log.Error(fmt.Sprintf("获取版本列表失败: %v", err))
		return ""
	}

	// 查找匹配的版本
	var targetRelease *DotNetReleaseDetail
	for i, release := range releases {
		if release.ReleaseVersion == version {
			targetRelease = &releases[i]
			break
		}
	}

	if targetRelease == nil {
		utils.Log.Warning(fmt.Sprintf("未找到版本 %s", version))
		return ""
	}

	// 根据组件类型获取对应的文件
	var files []DotNetComponentFile
	switch p.componentType {
	case "sdk":
		files = targetRelease.SDK.Files
	case "runtime":
		files = targetRelease.Runtime.Files
	case "asp-core":
		files = targetRelease.AspNetCore.Files
	case "desktop":
		files = targetRelease.WindowsDesktop.Files
	}

	// 如果组件没有单独的Files字段，则使用总的Files字段
	if len(files) == 0 {
		files = targetRelease.Files
	}

	// 构建RID（Runtime Identifier）
	var rid string
	switch osName {
	case "windows":
		if arch == "amd64" {
			rid = "win-x64"
		} else if arch == "386" {
			rid = "win-x86"
		} else if arch == "arm64" {
			rid = "win-arm64"
		}
	case "darwin":
		if arch == "amd64" {
			rid = "osx-x64"
		} else if arch == "arm64" {
			rid = "osx-arm64"
		}
	case "linux":
		if arch == "amd64" {
			rid = "linux-x64"
		} else if arch == "386" {
			rid = "linux-x86"
		} else if arch == "arm64" {
			rid = "linux-arm64"
		} else if arch == "arm" {
			rid = "linux-arm"
		}
	}

	// 如果没有指定架构，默认使用当前系统架构
	if rid == "" {
		switch runtime.GOOS {
		case "windows":
			if runtime.GOARCH == "amd64" {
				rid = "win-x64"
			} else if runtime.GOARCH == "386" {
				rid = "win-x86"
			} else if runtime.GOARCH == "arm64" {
				rid = "win-arm64"
			}
		case "darwin":
			if runtime.GOARCH == "amd64" {
				rid = "osx-x64"
			} else if runtime.GOARCH == "arm64" {
				rid = "osx-arm64"
			}
		default:
			if runtime.GOARCH == "amd64" {
				rid = "linux-x64"
			} else if runtime.GOARCH == "386" {
				rid = "linux-x86"
			} else if runtime.GOARCH == "arm64" {
				rid = "linux-arm64"
			} else if runtime.GOARCH == "arm" {
				rid = "linux-arm"
			}
		}
	}

	utils.Log.Search(fmt.Sprintf("查找适用于 %s 的 %s %s 下载", rid, p.componentType, version))

	// 首先尝试查找精确匹配的文件
	var bestMatch string
	var bestMatchScore int = -1

	for _, file := range files {
		// 检查是否为支持的安装包格式（排除exe文件）
		isSupported := strings.HasSuffix(file.Name, ".zip") ||
			strings.HasSuffix(file.Name, ".tar.gz") ||
			strings.HasSuffix(file.Name, ".pkg")

		if !isSupported {
			continue // 跳过不支持的格式，包括exe文件
		}

		// 计算匹配分数
		score := 0

		// 检查是否匹配当前平台
		if strings.Contains(file.Name, rid) {
			score += 10
		} else {
			continue // 必须匹配平台
		}

		// 检查是否匹配组件类型
		switch p.componentType {
		case "sdk":
			if strings.Contains(file.Name, "sdk") {
				score += 5
			}
		case "runtime":
			if strings.Contains(file.Name, "runtime") &&
				!strings.Contains(file.Name, "aspnetcore") &&
				!strings.Contains(file.Name, "windowsdesktop") {
				score += 5
			}
		case "asp-core":
			if strings.Contains(file.Name, "aspnetcore") {
				score += 5
			}
		case "desktop":
			if strings.Contains(file.Name, "windowsdesktop") {
				score += 5
			}
		}

		// 检查是否匹配版本
		if strings.Contains(file.Name, version) {
			score += 3
		}

		// 检查是否为首选的安装包格式
		switch osName {
		case "windows":
			if strings.HasSuffix(file.Name, ".zip") {
				score += 5
			}
		case "darwin":
			if strings.HasSuffix(file.Name, ".pkg") {
				score += 5
			} else if strings.HasSuffix(file.Name, ".tar.gz") {
				score += 3
			}
		default:
			if strings.HasSuffix(file.Name, ".tar.gz") {
				score += 5
			}
		}

		// 更新最佳匹配
		if score > bestMatchScore {
			bestMatchScore = score
			bestMatch = file.URL
			utils.Log.Info(fmt.Sprintf("找到更好的匹配: %s (分数: %d)", file.Name, score))
		}
	}

	if bestMatch != "" {
		utils.Log.Download(fmt.Sprintf("已找到下载链接: %s", bestMatch))
		return bestMatch
	}

	utils.Log.Warning(fmt.Sprintf("未找到适用于 %s-%s 的 %s %s 下载", osName, arch, p.componentType, version))
	return ""
}

// GetExtractDir 实现SDKProvider接口，获取解压后的目录名
func (p *DotNetSDKProvider) GetExtractDir(version, downloadedFile string) string {
	// 返回空字符串，表示不需要移动文件
	// 我们将在PostInstall中处理目录结构
	return ""
}

// GetBinDir 实现SDKProvider接口，获取bin目录
func (p *DotNetSDKProvider) GetBinDir(baseDir string) string {
	return baseDir
}

// ConfigureEnv 实现SDKProvider接口，配置环境变量
func (p *DotNetSDKProvider) ConfigureEnv(version, installDir string) ([]config.EnvVar, error) {
	// 构建组件目录和组件内的 current 目录
	baseDir := filepath.Join(installDir, "..")
	componentDir := filepath.Join(baseDir, p.componentType)
	currentDir := filepath.Join(componentDir, "current")

	utils.Log.Config(fmt.Sprintf("配置环境变量，组件目录: %s", componentDir))
	utils.Log.Config(fmt.Sprintf("组件内的 current 目录: %s", currentDir))

	// 设置环境变量
	var envVars []config.EnvVar

	if p.componentType == "sdk" {
		// 添加DOTNET_ROOT环境变量，指向current目录
		envVars = append(envVars, config.EnvVar{
			Key:   "DOTNET_ROOT",
			Value: currentDir,
		})
	}

	// 添加PATH环境变量，使用组件内的 current 目录
	envVars = append(envVars, config.EnvVar{
		Key:   "PATH",
		Value: currentDir,
	})

	// 添加排除关键字环境变量，避免PATH重复累加
	// 这将告诉 EnvManager 从 PATH 中移除包含这些关键字的路径
	envVars = append(envVars, config.EnvVar{
		Key:   "EXCLUDE_KEYWORDS",
		Value: fmt.Sprintf("%s,%s", baseDir, filepath.Join(baseDir, p.componentType)),
	})

	return envVars, nil
}

// PreInstall 实现SDKProvider接口，安装前的准备工作
func (p *DotNetSDKProvider) PreInstall(version string) error {
	return nil
}

// PostInstall 实现SDKProvider接口，安装后的处理工作
func (p *DotNetSDKProvider) PostInstall(version, installDir string) error {
	// 构建正确的目录结构 installDir/componentType/version
	baseDir := filepath.Join(installDir, "..")
	componentDir := filepath.Join(baseDir, p.componentType)
	versionDir := filepath.Join(componentDir, version)

	utils.Log.Info(fmt.Sprintf("基础目录: %s", baseDir))
	utils.Log.Info(fmt.Sprintf("组件目录: %s", componentDir))
	utils.Log.Info(fmt.Sprintf("版本目录: %s", versionDir))

	// 确保组件目录存在
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		return fmt.Errorf("创建组件目录失败: %w", err)
	}

	// 确保版本目录存在
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("创建版本目录失败: %w", err)
	}

	// 检查文件是否解压到了baseDir/version目录
	incorrectVersionDir := filepath.Join(baseDir, version)
	if _, err := os.Stat(incorrectVersionDir); err == nil {
		utils.Log.Warning(fmt.Sprintf("文件被解压到了错误的目录: %s，需要移动到正确的目录结构", incorrectVersionDir))

		// 移动所有文件到正确的版本目录
		entries, err := os.ReadDir(incorrectVersionDir)
		if err != nil {
			return fmt.Errorf("读取目录失败: %w", err)
		}

		for _, entry := range entries {
			src := filepath.Join(incorrectVersionDir, entry.Name())
			dst := filepath.Join(versionDir, entry.Name())

			// 如果目标文件已存在，先删除
			if _, err := os.Stat(dst); err == nil {
				if err := os.RemoveAll(dst); err != nil {
					utils.Log.Warning(fmt.Sprintf("无法删除已存在的文件: %s, 错误: %v", dst, err))
					continue
				}
			}

			// 移动文件
			utils.Log.Move(fmt.Sprintf("移动文件: %s -> %s", src, dst))
			if err := os.Rename(src, dst); err != nil {
				// 如果移动失败，尝试复制
				if entry.IsDir() {
					if err := utils.CopyDir(src, dst); err != nil {
						utils.Log.Warning(fmt.Sprintf("复制目录失败: %s -> %s, 错误: %v", src, dst, err))
					} else {
						os.RemoveAll(src) // 复制成功后删除源目录
					}
				} else {
					if err := utils.CopyFile(src, dst); err != nil {
						utils.Log.Warning(fmt.Sprintf("复制文件失败: %s -> %s, 错误: %v", src, dst, err))
					} else {
						os.Remove(src) // 复制成功后删除源文件
					}
				}
			}
		}

		// 删除空的版本目录
		if err := os.RemoveAll(incorrectVersionDir); err != nil {
			utils.Log.Warning(fmt.Sprintf("删除空目录失败: %s, 错误: %v", incorrectVersionDir, err))
		}
	} else {
		// 检查文件是否直接解压到了baseDir
		dotnetExeInBase := filepath.Join(baseDir, "dotnet.exe")
		if _, err := os.Stat(dotnetExeInBase); err == nil {
			utils.Log.Warning("文件被解压到了基础目录，需要移动到正确的目录结构")

			// 移动所有文件到版本目录
			entries, err := os.ReadDir(baseDir)
			if err != nil {
				return fmt.Errorf("读取基础目录失败: %w", err)
			}

			for _, entry := range entries {
				// 跳过组件目录本身
				if entry.Name() == p.componentType {
					continue
				}

				src := filepath.Join(baseDir, entry.Name())
				dst := filepath.Join(versionDir, entry.Name())

				// 如果目标文件已存在，先删除
				if _, err := os.Stat(dst); err == nil {
					if err := os.RemoveAll(dst); err != nil {
						utils.Log.Warning(fmt.Sprintf("无法删除已存在的文件: %s, 错误: %v", dst, err))
						continue
					}
				}

				// 移动文件
				utils.Log.Move(fmt.Sprintf("移动文件: %s -> %s", src, dst))
				if err := os.Rename(src, dst); err != nil {
					// 如果移动失败，尝试复制
					if entry.IsDir() {
						if err := utils.CopyDir(src, dst); err != nil {
							utils.Log.Warning(fmt.Sprintf("复制目录失败: %s -> %s, 错误: %v", src, dst, err))
						} else {
							os.RemoveAll(src) // 复制成功后删除源目录
						}
					} else {
						if err := utils.CopyFile(src, dst); err != nil {
							utils.Log.Warning(fmt.Sprintf("复制文件失败: %s -> %s, 错误: %v", src, dst, err))
						} else {
							os.Remove(src) // 复制成功后删除源文件
						}
					}
				}
			}
		}
	}

	// 检查dotnet可执行文件是否存在
	dotnetExe := filepath.Join(versionDir, "dotnet.exe")
	utils.Log.Check(fmt.Sprintf("检查dotnet可执行文件: %s", dotnetExe))

	if _, err := os.Stat(dotnetExe); os.IsNotExist(err) {
		// 如果在预期位置找不到，尝试在整个目录中查找
		var foundDotnetExe string

		// 在整个安装目录中查找
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && filepath.Base(path) == "dotnet.exe" {
				foundDotnetExe = path
				return filepath.SkipDir
			}
			return nil
		})

		if err != nil {
			utils.Log.Warning(fmt.Sprintf("在安装目录中查找dotnet.exe时出错: %v", err))
		}

		if foundDotnetExe != "" {
			utils.Log.Success(fmt.Sprintf("在其他位置找到dotnet可执行文件: %s", foundDotnetExe))

			// 如果在其他位置找到，尝试复制到预期位置
			if err := utils.CopyFile(foundDotnetExe, dotnetExe); err != nil {
				utils.Log.Warning(fmt.Sprintf("复制dotnet可执行文件失败: %v", err))
				// 使用找到的路径
				dotnetExe = foundDotnetExe
			}

			// 如果dotnet.exe在其他目录中，尝试移动整个目录的内容
			foundDir := filepath.Dir(foundDotnetExe)
			if foundDir != versionDir {
				utils.Log.Move(fmt.Sprintf("移动目录内容: %s -> %s", foundDir, versionDir))

				entries, err := os.ReadDir(foundDir)
				if err != nil {
					return fmt.Errorf("读取目录失败: %w", err)
				}

				for _, entry := range entries {
					src := filepath.Join(foundDir, entry.Name())
					dst := filepath.Join(versionDir, entry.Name())

					// 如果目标文件已存在，先删除
					if _, err := os.Stat(dst); err == nil {
						if err := os.RemoveAll(dst); err != nil {
							utils.Log.Warning(fmt.Sprintf("无法删除已存在的文件: %s, 错误: %v", dst, err))
							continue
						}
					}

					// 移动文件
					utils.Log.Move(fmt.Sprintf("移动文件: %s -> %s", src, dst))
					if err := os.Rename(src, dst); err != nil {
						// 如果移动失败，尝试复制
						if entry.IsDir() {
							if err := utils.CopyDir(src, dst); err != nil {
								utils.Log.Warning(fmt.Sprintf("复制目录失败: %s -> %s, 错误: %v", src, dst, err))
							} else {
								os.RemoveAll(src) // 复制成功后删除源目录
							}
						} else {
							if err := utils.CopyFile(src, dst); err != nil {
								utils.Log.Warning(fmt.Sprintf("复制文件失败: %s -> %s, 错误: %v", src, dst, err))
							} else {
								os.Remove(src) // 复制成功后删除源文件
							}
						}
					}
				}

				// 尝试删除源目录（如果不是基础目录或组件目录）
				if foundDir != baseDir && foundDir != componentDir && foundDir != versionDir {
					if err := os.RemoveAll(foundDir); err != nil {
						utils.Log.Warning(fmt.Sprintf("删除源目录失败: %s, 错误: %v", foundDir, err))
					}
				}
			}
		} else {
			return fmt.Errorf("安装可能不完整，找不到dotnet可执行文件")
		}
	}

	// 设置可执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dotnetExe, 0755); err != nil {
			return fmt.Errorf("设置可执行权限失败: %w", err)
		}
	}

	utils.Log.Success(fmt.Sprintf("已成功安装 .NET %s %s", p.componentType, version))
	return nil
}

// GetArchiveType 实现SDKProvider接口，获取归档类型
func (p *DotNetSDKProvider) GetArchiveType() string {
	return "auto"
}

// GetArchiveTypeForFile 实现SDKProvider接口，根据具体文件确定归档类型
func (p *DotNetSDKProvider) GetArchiveTypeForFile(filePath string) string {
	utils.Log.Info(fmt.Sprintf("检测文件类型: %s", filePath))

	if strings.HasSuffix(filePath, ".zip") {
		utils.Log.Extract("检测到zip文件")
		return "zip"
	} else if strings.HasSuffix(filePath, ".tar.gz") {
		utils.Log.Extract("检测到tar.gz文件")
		return "tar.gz"
	} else if strings.HasSuffix(filePath, ".pkg") {
		utils.Log.Extract("检测到pkg文件")
		return "pkg"
	}
	// 不再处理.exe文件
	utils.Log.Warning(fmt.Sprintf("未知文件类型: %s", filePath))
	return "unknown"
}
