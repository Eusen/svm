package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// EnvVar 表示环境变量
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SDKVersionInfo 表示SDK版本信息
type SDKVersionInfo struct {
	InstallDir    string `json:"install_dir"`
	CacheFilePath string `json:"cache_file_path"`
}

// SDKConfig 表示单个SDK的配置
type SDKConfig struct {
	CurrentVersion string                    `json:"current_version"`
	EnvVars        []EnvVar                  `json:"env_vars"`
	VersionCache   map[string]SDKVersionInfo `json:"version_cache"`
}

// Config 表示全局配置
type Config struct {
	InstallDir      string               `json:"install_dir"`
	CurrentVersions map[string]string    `json:"current_versions"` // 为向后兼容保留
	SDKs            map[string]SDKConfig `json:"sdks"`             // 新增SDK配置
}

func GetDefaultInstallDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".svm")
	}
	return filepath.Join(homeDir, ".svm")
}

func LoadConfig() (*Config, error) {
	configFile := getConfigFilePath()

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := &Config{
			InstallDir:      GetDefaultInstallDir(),
			CurrentVersions: make(map[string]string),
			SDKs:            make(map[string]SDKConfig),
		}
		return cfg, cfg.Save()
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 确保各种map已初始化
	if cfg.CurrentVersions == nil {
		cfg.CurrentVersions = make(map[string]string)
	}

	if cfg.SDKs == nil {
		cfg.SDKs = make(map[string]SDKConfig)
	}

	// 如果InstallDir为空，使用默认值
	if cfg.InstallDir == "" {
		cfg.InstallDir = GetDefaultInstallDir()
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	configFile := getConfigFilePath()
	configDir := filepath.Dir(configFile)

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func (c *Config) SetInstallDir(dir string) error {
	c.InstallDir = dir
	return c.Save()
}

func (c *Config) GetCurrentVersion(sdk string) string {
	if sdkConfig, ok := c.SDKs[sdk]; ok && sdkConfig.CurrentVersion != "" {
		return sdkConfig.CurrentVersion
	}
	return c.CurrentVersions[sdk]
}

func (c *Config) SetCurrentVersion(sdk, version string) error {
	// 更新旧的配置
	c.CurrentVersions[sdk] = version

	// 更新新的配置
	if _, ok := c.SDKs[sdk]; !ok {
		c.SDKs[sdk] = SDKConfig{
			VersionCache: make(map[string]SDKVersionInfo),
		}
	}

	// 创建一个临时变量修改，然后重新赋值
	sdkConfig := c.SDKs[sdk]
	sdkConfig.CurrentVersion = version
	c.SDKs[sdk] = sdkConfig

	return c.Save()
}

func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".svm", "config.json")
	}
	return filepath.Join(homeDir, ".svm", "config.json")
}

// GetCurrentVersionInfo 获取指定SDK的特定版本信息
func (c *Config) GetVersionInfo(sdk, version string) (SDKVersionInfo, bool) {
	sdkConfig, ok := c.SDKs[sdk]
	if !ok {
		return SDKVersionInfo{}, false
	}

	if sdkConfig.VersionCache == nil {
		return SDKVersionInfo{}, false
	}

	info, ok := sdkConfig.VersionCache[version]
	return info, ok
}

// SetVersionInfo 设置版本信息
func (c *Config) SetVersionInfo(sdk, version string, info SDKVersionInfo) error {
	// 确保SDK配置存在
	if _, ok := c.SDKs[sdk]; !ok {
		c.SDKs[sdk] = SDKConfig{
			VersionCache: make(map[string]SDKVersionInfo),
		}
	}

	// 创建一个临时变量修改，然后重新赋值
	sdkConfig := c.SDKs[sdk]

	// 确保VersionCache已初始化
	if sdkConfig.VersionCache == nil {
		sdkConfig.VersionCache = make(map[string]SDKVersionInfo)
	}

	// 设置版本信息
	sdkConfig.VersionCache[version] = info
	c.SDKs[sdk] = sdkConfig

	// 同时更新原有的CurrentVersions（为了向后兼容）
	if sdkConfig.CurrentVersion != "" {
		c.CurrentVersions[sdk] = sdkConfig.CurrentVersion
	}

	return c.Save()
}

// GetSDKEnvVars 获取SDK的环境变量
func (c *Config) GetSDKEnvVars(sdk string) []EnvVar {
	sdkConfig, ok := c.SDKs[sdk]
	if !ok || sdkConfig.EnvVars == nil {
		return []EnvVar{}
	}
	return sdkConfig.EnvVars
}

// SetSDKEnvVars 设置SDK的环境变量
func (c *Config) SetSDKEnvVars(sdk string, envVars []EnvVar) error {
	// 确保SDK配置存在
	if _, ok := c.SDKs[sdk]; !ok {
		c.SDKs[sdk] = SDKConfig{
			VersionCache: make(map[string]SDKVersionInfo),
		}
	}

	// 创建一个临时变量修改，然后重新赋值
	sdkConfig := c.SDKs[sdk]
	sdkConfig.EnvVars = envVars
	c.SDKs[sdk] = sdkConfig

	return c.Save()
}

// GetCacheDir 返回缓存目录路径
func (c *Config) GetCacheDir() string {
	return filepath.Join(c.InstallDir, "cache")
}

// RemoveVersionInfo 从配置中移除指定SDK的指定版本信息
func (c *Config) RemoveVersionInfo(sdk, version string) error {
	// 检查SDK配置是否存在
	sdkConfig, ok := c.SDKs[sdk]
	if !ok {
		// SDK不存在，无需移除
		return nil
	}

	// 检查VersionCache是否已初始化
	if sdkConfig.VersionCache == nil {
		// VersionCache不存在，无需移除
		return nil
	}

	// 检查版本是否存在
	if _, exists := sdkConfig.VersionCache[version]; !exists {
		// 版本不存在，无需移除
		return nil
	}

	// 移除版本信息
	delete(sdkConfig.VersionCache, version)
	c.SDKs[sdk] = sdkConfig

	// 如果当前版本是被移除的版本，清空当前版本
	if sdkConfig.CurrentVersion == version {
		sdkConfig.CurrentVersion = ""
		c.SDKs[sdk] = sdkConfig
		delete(c.CurrentVersions, sdk)
	}

	return c.Save()
}
